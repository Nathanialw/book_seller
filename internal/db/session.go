package db

import "github.com/gorilla/sessions"

var Store = sessions.NewCookieStore([]byte("super-secret-key")) // use a strong key in production
