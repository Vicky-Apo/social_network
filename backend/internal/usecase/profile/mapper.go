package profile

import domainuser "social-network/backend/internal/domain/user"

func mapUsers(users []domainuser.User) []UserDTO {
	out := make([]UserDTO, 0, len(users))
	for _, u := range users {
		out = append(out, mapUser(u))
	}
	return out
}

func mapUser(u domainuser.User) UserDTO {
	return UserDTO{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DateOfBirth: u.DateOfBirth.Format("02/01/2006"),
		AvatarPath:  u.AvatarPath,
		Nickname:    u.Nickname,
		About:       u.About,
		IsPublic:    u.IsPublic,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func mapUserLimited(u domainuser.User) LimitedUserDTO {
	return LimitedUserDTO{
		ID:         u.ID,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Nickname:   u.Nickname,
		AvatarPath: u.AvatarPath,
		IsPublic:   u.IsPublic,
	}
}

func mapProfileUser(u domainuser.User) ProfileUserDTO {
	email := u.Email
	dob := u.DateOfBirth.Format("02/01/2006")
	created := u.CreatedAt
	updated := u.UpdatedAt
	return ProfileUserDTO{
		ID:          u.ID,
		Email:       &email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		DateOfBirth: &dob,
		AvatarPath:  u.AvatarPath,
		Nickname:    u.Nickname,
		About:       u.About,
		IsPublic:    u.IsPublic,
		CreatedAt:   &created,
		UpdatedAt:   &updated,
	}
}

func mapProfileUserLimited(u domainuser.User) ProfileUserDTO {
	return ProfileUserDTO{
		ID:         u.ID,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Nickname:   u.Nickname,
		AvatarPath: u.AvatarPath,
		IsPublic:   u.IsPublic,
	}
}
