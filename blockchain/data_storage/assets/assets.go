package assets

import (
	"errors"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Assets struct {
	hash_map.HashMap `json:"-"`
}

func (assets *Assets) GetAsset(key []byte) (*asset.Asset, error) {

	if len(key) == 0 {
		key = config_coins.NATIVE_ASSET_FULL
	}

	data, err := assets.HashMap.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*asset.Asset), nil
}

func (assets *Assets) CreateAsset(key []byte, ast *asset.Asset) (err error) {

	if len(key) == 0 {
		key = config_coins.NATIVE_ASSET_FULL
	}

	if err = ast.Validate(); err != nil {
		return
	}

	var exists bool
	if exists, err = assets.ExistsAsset(key); err != nil {
		return
	}
	if exists {
		return errors.New("Asset already exists")
	}

	if err = assets.UpdateAsset(key, ast); err != nil {
		return
	}
	return
}

func (assets *Assets) UpdateAsset(key []byte, ast *asset.Asset) error {

	if len(key) == 0 {
		key = config_coins.NATIVE_ASSET_FULL
	}

	return assets.Update(string(key), ast)
}

func (assets *Assets) ExistsAsset(key []byte) (bool, error) {

	if len(key) == 0 {
		key = config_coins.NATIVE_ASSET_FULL
	}

	return assets.Exists(string(key))
}

func (assets *Assets) DeleteAsset(key []byte) {
	assets.Delete(string(key))
}

func NewAssets(tx store_db_interface.StoreDBTransactionInterface) (assets *Assets) {

	hashMap := hash_map.CreateNewHashMap(tx, "assets", config_coins.ASSET_LENGTH, true)

	assets = &Assets{
		HashMap: *hashMap,
	}
	assets.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var ast = &asset.Asset{}
		if err := ast.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
		return ast, nil
	}
	return
}
