package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yusufdot101/goBankBackend/internal/user"
)

var app Application

func TestSetUserContext(t *testing.T) {
	mockUser := &user.User{
		ID:    1,
		Name:  "yusuf",
		Email: "ym@gmail.com",
	}
	buf := new(bytes.Buffer)
	req := httptest.NewRequest(http.MethodGet, "/", buf)
	req = app.setUserContext(req, mockUser)
	gotUser, ok := req.Context().Value(userContextKey).(*user.User)
	if !ok {
		t.Fatal("user key missing in request context")
	}

	if gotUser.ID != mockUser.ID {
		t.Errorf("expected user id=%d, got id=%d", mockUser.ID, gotUser.ID)
	}
	if gotUser.Name != mockUser.Name {
		t.Errorf("expected user name=%s, got name=%s", mockUser.Name, gotUser.Name)
	}
	if gotUser.Email != mockUser.Email {
		t.Errorf("expected user email=%s, got email=%s", mockUser.Email, gotUser.Email)
	}
}

func TestGetUserContext(t *testing.T) {
	mockUser := &user.User{
		ID:    1,
		Name:  "yusuf",
		Email: "ym@gmail.com",
	}
	buf := new(bytes.Buffer)
	req := httptest.NewRequest(http.MethodGet, "/", buf)
	ctx := context.WithValue(req.Context(), userContextKey, mockUser)
	req = req.WithContext(ctx)

	gotUser := app.getUserContext(req)
	if gotUser.ID != mockUser.ID {
		t.Errorf("expected user id=%d, got id=%d", mockUser.ID, gotUser.ID)
	}
	if gotUser.Name != mockUser.Name {
		t.Errorf("expected user name=%s, got name=%s", mockUser.Name, gotUser.Name)
	}
	if gotUser.Email != mockUser.Email {
		t.Errorf("expected user email=%s, got email=%s", mockUser.Email, gotUser.Email)
	}
}
