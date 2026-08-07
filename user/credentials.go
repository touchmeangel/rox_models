package user

type Credentials struct {
	User
	PasswordHash string
}

type credentialsRow struct {
	ID           string   `db:"id"`
	Email        string   `db:"email"`
	Username     string   `db:"username"`
	Roles        []string `db:"roles"`
	PasswordHash string   `db:"password_hash"`
}

func (r credentialsRow) toSDK() Credentials {
	roles := make([]Role, len(r.Roles))
	for i, ro := range r.Roles {
		roles[i] = Role(ro)
	}
	return Credentials{
		User: User{
			ID:       r.ID,
			Email:    r.Email,
			Username: r.Username,
			Roles:    roles,
		},
		PasswordHash: r.PasswordHash,
	}
}
