package main

import "errors"

var USER_EXISTS = errors.New("User already exists")
var USER_DOESNT_EXIST = errors.New("User does not exist")
var INVALID_LOGIN = errors.New("Invalid login")
var INVALID_SESSION = errors.New("Invalid session")
