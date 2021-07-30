package authserv2

import "errors"

// declare errors
var (
	ErrDBoffline = errors.New("database is offline")

	// login
	ErrLoginNotFound     = errors.New("login not found")
	ErrPasswordMissmatch = errors.New("password missmatch")

	// register
	ErrLoginInUse        = errors.New("login is in use")
	ErrLoginRequirements = errors.New("login doesnt match requirements")
	ErrPassRequirements  = errors.New("password doesnt match requirements")

	ErrUserNotLogedIn  = errors.New("user is loged out, logged_in=false")
	ErrTokenExpiration = errors.New("token has expired")
	ErrTokenMissmatch  = errors.New("tokens dont match")
)
