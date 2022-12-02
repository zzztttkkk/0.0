package account

import "context"

type RegisterInput struct {
	Email    string `vld:"email;regexp=email"`
	Nickname string `vld:"nickname;nullable"`
}

type RegisterOut struct {
	Id int64 `json:"id"`
}

//@security Un
//@expose http post /api/account/register
func register(ctx context.Context, params RegisterInput) (*RegisterOut, error) {
	return nil, nil
}
