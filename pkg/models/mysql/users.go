package mysql

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/yoqub-davlatov/snippetbox/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

const (
	passwordCostEncryption     = 12
	duplicateEntryError        = 1062
	usersUniqueEmailConstraint = "users_uc_email"
)

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), passwordCostEncryption)
	if err != nil {
		return err
	}

	stmt := `
		INSERT INTO users (name, email, hashed_password, created)
		VALUES(?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))

	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == duplicateEntryError &&
				strings.Contains(mySQLError.Message, usersUniqueEmailConstraint) {
				return models.ErrDuplicateEmail
			}
		}
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	return 0, nil
}

func (m *UserModel) Get(id int) (*models.User, error) {
	return nil, nil
}
