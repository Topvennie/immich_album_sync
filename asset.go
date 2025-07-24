package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

type Asset struct {
	ID   string `json:"id"`
	Name string `json:"originalPath"`
}

func (u *User) getImmichAssets(ctx context.Context) ([]Asset, error) {
	type body struct {
		Page int `json:"page"`
		Size int `json:"size"`
	}

	type response struct {
		Assets struct {
			Count int     `json:"count"`
			Items []Asset `json:"items"`
		} `json:"assets"`
	}

	assets := []Asset{}

	page := 1
	size := 1000

	for {
		data := body{
			Page: page,
			Size: size,
		}

		buf, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		resp, err := u.immichRequest(ctx, "POST", "search/metadata", bytes.NewReader(buf))
		if err != nil {
			return nil, err
		}

		var newAssets response
		if err := json.Unmarshal(resp, &newAssets); err != nil {
			return nil, err
		}

		assets = append(assets, newAssets.Assets.Items...)
		if newAssets.Assets.Count != size {
			break
		}

		page++
	}

	return assets, nil
}

func (u *User) getAlbumAssets(ctx context.Context, album *Album) error {
	data, err := u.immichRequest(ctx, "GET", "albums/"+album.ID, nil)
	if err != nil {
		return err
	}

	var fullAlbum Album
	if err := json.Unmarshal(data, &fullAlbum); err != nil {
		return err
	}

	*album = fullAlbum

	return nil
}

func (u *User) AddAlbumAssets(ctx context.Context, album *Album, assets []Asset) error {
	type body struct {
		IDs []string `json:"ids"`
	}

	data := body{
		IDs: make([]string, 0, len(assets)),
	}
	for _, asset := range assets {
		data.IDs = append(data.IDs, asset.ID)
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := u.immichRequest(ctx, "PUT", "albums/"+album.ID+"/assets", bytes.NewReader(buf)); err != nil {
		return err
	}

	album.Assets = append(album.Assets, assets...)

	return nil
}

func (u *User) RemoveAlbumAssets(ctx context.Context, album *Album, assets []Asset) error {
	type body struct {
		IDs []string `json:"ids"`
	}

	data := body{
		IDs: make([]string, 0, len(assets)),
	}
	for _, asset := range assets {
		data.IDs = append(data.IDs, asset.ID)
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if _, err := u.immichRequest(ctx, "DELETE", "albums/"+album.ID+"/assets", bytes.NewReader(buf)); err != nil {
		return err
	}

	assetMap := make(map[string]Asset, len(assets))
	for _, asset := range assets {
		assetMap[asset.ID] = asset
	}

	newAssets := []Asset{}
	for _, asset := range album.Assets {
		if _, ok := assetMap[asset.ID]; !ok {
			newAssets = append(newAssets, asset)
		}
	}

	album.Assets = newAssets

	return nil
}

func (u *User) syncAssets(ctx context.Context, assets []Asset, external, immich []Album) error {
	immichMap := make(map[string]Album, len(immich))
	for _, album := range immich {
		immichMap[album.Name] = album
	}
	assetMap := make(map[string]Asset, len(assets))
	for _, asset := range assets {
		assetMap[asset.Name] = asset
	}

	for _, e := range external {
		album, ok := immichMap[e.Name]
		if !ok {
			continue
		}

		// Add assets
		albumAssetMap := make(map[string]Asset, len(album.Assets))
		for _, asset := range album.Assets {
			albumAssetMap[asset.Name] = asset
		}

		var addAssets []Asset
		for _, asset := range e.Assets {
			if _, ok := assetMap[asset.Name]; !ok {
				continue // Image is not in immich yet
			}
			if _, ok := albumAssetMap[asset.Name]; !ok {
				addAssets = append(addAssets, asset)
			}
		}

		if addAssets != nil {
			if err := u.AddAlbumAssets(ctx, &album, addAssets); err != nil {
				return err
			}
		}

		// Remove assets
		externalAssetMap := make(map[string]Asset, len(e.Assets))
		for _, asset := range e.Assets {
			externalAssetMap[asset.Name] = asset
		}
		fmt.Println(externalAssetMap)

		var removeAssets []Asset
		for _, asset := range album.Assets {
			if _, ok := externalAssetMap[asset.Name]; !ok {
				removeAssets = append(removeAssets, asset)
			}
		}

		if removeAssets != nil {
			if err := u.RemoveAlbumAssets(ctx, &album, removeAssets); err != nil {
				return err
			}
		}
	}

	return nil
}
