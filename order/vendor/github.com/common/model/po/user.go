package po

type UserInfoPo struct {
	Id       int64  `db:"id"`
	Name     string `db:"name"`
	NickName string `db:"nickname"`
	Gender   int    `db:"gender"`
	Age      int    `db:"age"`
}

type UserPasswordPo struct {
	UserId   int64  `db:"user_id"`
	Password string `db:"encrpt_password"`
}
