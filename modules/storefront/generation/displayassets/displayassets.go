package displayassets

import "fmt"

func SetDisplayAsset(item string) string {
	return fmt.Sprintf("/Game/Catalog/DisplayAssets/%s.%s", item, item)
}

func SetNewDisplayAssetPath(item string) string {
	return fmt.Sprintf("/Game/Catalog/NewDisplayAssets/DAv2_Featured_%s.DAv2_Featured_%s", item, item)
}
