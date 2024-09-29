package storage

type UsersChats struct {
	Users map[string][]string
}

func NewUsersChats() *UsersChats {
	return &UsersChats{
		Users: map[string][]string{
			"user1": {"chat1", "chat2", "chat3"},
			"user2": {"chat1", "chat2", "chat3"},
		},
	}
}
