package auth

// User is a type that represents a single user as it is stored in the database
type User struct {
	ID           uint64 `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"_"`
}

// UserRegister represents the data that we expect to recieve from the
// registration page.
type UserRegister struct {
	Username   string `json:"username" form:"username" binding:"required"`
	Email      string `json:"email" form:"email" binding:"required"`
	Password   string `json:"password" form:"password" binding:"required"`
	RePassword string `json:"repassword" form:"repassword" binding:"required"`
}

// UserLogin represents the data that we expect to recieve from the
// login page.
type UserLogin struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// UserService is the interface that defines methods to work with Users
type UserService interface {
	// Creates a new user.
	// Returns an error if user with the same username or the same email
	// already exist or if passwords do not match.
	Create(u UserRegister) error
	// Authenticates a user by validating that it exists and hash of the
	// provided password matches. On success returns a JWT token.
	Authenticate(u UserLogin) (string, error)
	// Validates given token for a given user.
	Validate(u User, t string) bool
}
