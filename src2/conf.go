package authserv2

import "time"

type Conf struct {
	TOKEN_LENGTH  int
	MAX_TOKEN_AGE time.Duration
}

var CONFIG = Conf{
	TOKEN_LENGTH:  40,
	MAX_TOKEN_AGE: time.Hour,
}
