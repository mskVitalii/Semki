package crypto

import (
	"semki/internal/model"
)

// TODO: encrypt sensitive fields
// user description - user by default thinks that everything secure and may store secrets in description "I'm working on Half-Life 3"

func EncryptUserFields(user model.User, passphrase string) (*model.User, error) {
	//key := lib.GenerateKey(passphrase)
	encryptedUser := model.User{
		Id:        user.Id,
		Email:     user.Email,
		Password:  user.Password,
		Providers: user.Providers,
		Status:    user.Status,
	}

	//for i, fav := range user.Favourites {
	//	userFavJSON, err := json.Marshal(fav)
	//	if err != nil {
	//		return nil, err
	//	}
	//	encryptedFav, err := lib.Encrypt(userFavJSON, key)
	//
	//	if err != nil {
	//		return nil, err
	//	}
	//	encryptedUser.Favourites[i] = encryptedFav
	//}

	//for i, home := range user.Homes {
	//	userHomeJSON, err := json.Marshal(home)
	//	if err != nil {
	//		return nil, err
	//	}
	//	encryptedHome, err := lib.Encrypt(userHomeJSON, key)
	//
	//	if err != nil {
	//		return nil, err
	//	}
	//	encryptedUser.Homes[i] = encryptedHome
	//}

	return &encryptedUser, nil
}

func DecryptUserFields(user model.User, passphrase string) (*model.User, error) {
	//key := lib.GenerateKey(passphrase)
	decryptedUser := model.User{
		Id:        user.Id,
		Email:     user.Email,
		Password:  user.Password,
		Providers: user.Providers,
		//Favourites: make([]model.UserFavourite, len(user.Favourites)),
		//Homes:      make([]model.UserHome, len(user.Homes)),
		Status: user.Status,
	}

	//for i, fav := range user.Favourites {
	//	decryptedFav, err := lib.Decrypt(fav, key)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	var userFav model.UserFavourite
	//	err = json.Unmarshal([]byte(decryptedFav), &userFav)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	decryptedUser.Favourites[i] = userFav
	//}
	//
	//for i, home := range user.Homes {
	//	decryptedHome, err := lib.Decrypt(home, key)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	var userHome model.UserHome
	//	err = json.Unmarshal([]byte(decryptedHome), &userHome)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	decryptedUser.Homes[i] = userHome
	//}

	return &decryptedUser, nil
}
