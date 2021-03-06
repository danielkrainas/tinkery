package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/danielkrainas/gobag/api/errcode"
	"github.com/danielkrainas/gobag/context"
	"github.com/danielkrainas/gobag/decouple/cqrs"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/danielkrainas/tinkersnest/api/v1"
	"github.com/danielkrainas/tinkersnest/commands"
	"github.com/danielkrainas/tinkersnest/queries"
	"github.com/danielkrainas/tinkersnest/storage"
)

func blogListDispatcher(ctx context.Context, r *http.Request) http.Handler {
	h := &blogHandler{
		Context: ctx,
	}

	return handlers.MethodHandler{
		"GET":  withTraceLogging("GetAllPosts", h.GetAllPosts),
		"POST": withTraceLogging("CreatePost", h.CreatePost),
	}
}

func postByNameDispatcher(ctx context.Context, r *http.Request) http.Handler {
	h := &blogHandler{
		Context: ctx,
	}

	return handlers.MethodHandler{
		"GET":    withTraceLogging("GetPost", h.GetPost),
		"DELETE": withTraceLogging("DeletePost", h.DeletePost),
		"PUT":    withTraceLogging("UpdatePost", h.UpdatePost),
	}
}

type blogHandler struct {
	context.Context
}

func (ctx *blogHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postName := acontext.GetStringValue(ctx, "vars.post_name")
	postRaw, err := cqrs.DispatchQuery(ctx, &queries.FindPost{
		Name: postName,
	})

	if err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	if postRaw == nil {
		ctx.Context = acontext.AppendError(ctx, v1.ErrorCodeResourceUnknown)
		return
	}

	post := postRaw.(*v1.Post)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	p := &v1.Post{}
	if err = json.Unmarshal(body, p); err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	post.Publish = p.Publish
	if p.Title != "" {
		post.Title = p.Title
	}

	if len(p.Content) > 0 {
		post.Content = p.Content
	}

	if err := cqrs.DispatchCommand(ctx, &commands.StorePost{New: false, Post: post}); err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx.Context, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	if err := v1.ServeJSON(w, post); err != nil {
		acontext.GetLogger(ctx).Errorf("error sending post json: %v", err)
	}
}

func (ctx *blogHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postName := acontext.GetStringValue(ctx, "vars.post_name")
	err := cqrs.DispatchCommand(ctx, &commands.DeletePost{postName})
	if err != nil {
		if err == storage.ErrNotFound {
			acontext.GetLogger(ctx).Error("post not found")
			ctx.Context = acontext.AppendError(ctx, v1.ErrorCodeResourceUnknown)
			return
		} else {
			acontext.GetLogger(ctx).Error(err)
			ctx.Context = acontext.AppendError(ctx, errcode.ErrorCodeUnknown.WithDetail(err))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (ctx *blogHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	postName := acontext.GetStringValue(ctx, "vars.post_name")
	post, err := cqrs.DispatchQuery(ctx, &queries.FindPost{
		Name: postName,
	})

	if err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	if post == nil {
		ctx.Context = acontext.AppendError(ctx, v1.ErrorCodeResourceUnknown)
		return
	}

	if err := v1.ServeJSON(w, post); err != nil {
		acontext.GetLogger(ctx).Errorf("error sending post json: %v", err)
	}
}

func (ctx *blogHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	p := &v1.Post{}
	if err = json.Unmarshal(body, p); err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	if err := cqrs.DispatchCommand(ctx, &commands.StorePost{New: true, Post: p}); err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx.Context, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	acontext.GetLoggerWithField(ctx, "post.name", p.Name).Infof("blog post %q created", p.Name)
	if err := v1.ServeJSON(w, p); err != nil {
		acontext.GetLogger(ctx).Errorf("error sending blog post json: %v", err)
	}
}

func (ctx *blogHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	q := &queries.SearchPosts{}
	routeName := mux.CurrentRoute(r).GetName()
	if routeName == v1.RouteNamePostsByUser {
		q.Author = &v1.Author{User: acontext.GetStringValue(ctx, "vars.user_name")}
	}

	posts, err := cqrs.DispatchQuery(ctx, q)
	if err != nil {
		acontext.GetLogger(ctx).Error(err)
		ctx.Context = acontext.AppendError(ctx.Context, errcode.ErrorCodeUnknown.WithDetail(err))
		return
	}

	if err := v1.ServeJSON(w, posts); err != nil {
		acontext.GetLogger(ctx).Errorf("error sending posts json: %v", err)
	}
}
