package crypto

import (
	"semki/internal/model"
	"semki/pkg/lib"
)

// encrypt sensitive fields
// user description - user by default thinks that everything secure and may store secrets in description "I'm working on Half-Life 3"

func EncryptUserFields(user model.User, passphrase string) (*model.User, error) {
	key := lib.GenerateKey(passphrase)
	encryptedDesc, err := lib.Encrypt([]byte(user.Semantic.Description), key)
	if err != nil {
		return nil, err
	}
	user.Semantic.Description = encryptedDesc
	return &user, nil
}

func DecryptUserFields(user model.User, passphrase string) (*model.User, error) {
	key := lib.GenerateKey(passphrase)
	decryptedDesc, err := lib.Decrypt(user.Semantic.Description, key)
	if err != nil {
		return nil, err
	}
	user.Semantic.Description = decryptedDesc
	return &user, nil
}
