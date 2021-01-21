package services

import "golang.org/x/oauth2"

type Authenticated interface {
	SetToken(*oauth2.Token)
}