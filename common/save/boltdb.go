package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
	"pkg.nimblebun.works/wordle-cli/common"
)

const (
	rootUsersBucket = "users"
)

var (
	// ErrNoSuchUser is returned if the user doesn't exist
	ErrNoSuchUser = errors.New("the requested user does not exist")
	// ErrInvalidUserID is returned if the user ID provided is not valid
	// such as blank or too long.
	ErrInvalidUserID = errors.New("the provided user id is invalid or blank")

	ErrNoSuchSaveFile = errors.New("the requested save file does not exist")
)

// Database represents the connection to a backend datastore
type Database struct {
	bolt *bolt.DB
}

// User represents a user in the database by their public key.
type User struct {
	// UserID is used to identify the user and is the authoritative ID of the user.
	// It's derived from a FNV64a hash of the public key.
	UserID uint64 `json:"user_id"`
	// SaveFile is the save state of the user's game.
	SaveFiles map[common.GameType]SaveFile `json:"savefiles"`
}

func NewStorage(path string) (*Database, error) {
	db, err := getStorage(path)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db Database) Load(id string, user uint64) (*SaveFile, error) {
	u, err := db.getUser(user)
	if err != nil {
		return nil, err
	}
	save, ok := u.SaveFiles[common.StringToGameType(id)]
	if !ok {
		return nil, ErrNoSuchSaveFile
	}
	return &save, nil
}

func (db Database) Save(save *SaveFile, id string, user uint64) error {
	var u User
	var err error

	u, err = db.getUser(user)
	if err != nil {
		if errors.Is(err, ErrNoSuchUser) {
			u = User{
				SaveFiles: make(map[common.GameType]SaveFile),
				UserID:    user,
			}
		} else {
			return err
		}
	}
	u.SaveFiles[common.StringToGameType(id)] = *save
	return db.writeUser(u)
}

// getStorage returns a pointer to the bold DB instance of the provided file on disk.
// Should only be called once and then passed to whatever needs to connect to the db.
func getStorage(path string) (*Database, error) {
	// Open the database file with a timeout of 1 second to prevent it from hanging forever
	// if another process is reading that same file.
	boltDB, err := bolt.Open(path, 0666, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	db := &Database{
		bolt: boltDB,
	}
	err = db.ensureRootUserBucket()
	if err != nil {
		// This should never happen as the root user bucket is a constant.
		// Would only happen if there was a file issue with connecting to the
		// database, which makes it a good smoke test at initialisation phase.
		return nil, err
	}
	return db, nil
}

// getUser retrieves a User from the provided username.
// If the user doesn't exist, err will be storage.ErrNoSuchUser
func (db *Database) getUser(key uint64) (User, error) {
	var user User

	err := db.bolt.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(rootUsersBucket))
		if root == nil {
			// This should never happen, the root user bucket should always exist.
			return errors.New("user root bucket not defined")
		}
		rawUser := root.Get([]byte(strconv.FormatUint(key, 10)))
		if rawUser == nil {
			// The user doesn't exist
			return ErrNoSuchUser
		}
		err := json.Unmarshal(rawUser, &user)
		if err != nil {
			return fmt.Errorf("failed to unmarshal user: %w", err)
		}
		return nil
	})
	if err != nil {
		return user, err
	}
	return user, nil
}

// writeUser writes the provided user to the database.
// Returns an error if the user.PubKey is empty or if something
// went wrong in the update transaction.
func (db *Database) writeUser(user User) error {
	fmt.Printf("saving user %d\n", user.UserID)
	// First let's check if the user has all the data we need.
	if user.UserID == 0 {
		return ErrInvalidUserID
	}

	// Now we can write the user to the database.
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte(rootUsersBucket))
		if root == nil {
			// This should never happen, the root user bucket should always exist.
			return errors.New("user root bucket not defined")
		}
		rawUser, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("failed to marshal user: %w", err)
		}
		err = root.Put([]byte(strconv.FormatUint(user.UserID, 10)), rawUser)
		if err != nil {
			return fmt.Errorf("failed to write user to database: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil

}

func (db *Database) ensureRootUserBucket() error {
	err := db.bolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(rootUsersBucket))
		if err != nil {
			// This should never happen as the bucket name is a constant.
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
