package asset

import (
	"github.com/peterbourgon/mergemap"
	"github.com/raystack/compass/core/user"
)

// patch appends asset with data from map. It mutates the asset itself.
func patchAsset(a *Asset, patchData map[string]interface{}) {
	a.URN = patchString("urn", patchData, a.URN)
	a.Type = Type(patchString("type", patchData, a.Type.String()))
	a.Service = patchString("service", patchData, a.Service)
	a.Name = patchString("name", patchData, a.Name)
	a.Description = patchString("description", patchData, a.Description)
	a.URL = patchString("url", patchData, a.URL)

	labels, exists := patchData["labels"]
	if exists {
		a.Labels = mergeLabels(a.Labels, buildLabels(labels))
	}
	owners, exists := patchData["owners"]
	if exists {
		a.Owners = mergeOwners(a.Owners, buildOwners(owners))
	}
	data, exists := patchData["data"]
	if exists {
		patchAssetData(a, data)
	}
}

// mergeLabels merges new labels into existing ones. New values override existing keys.
func mergeLabels(existing, patch map[string]string) map[string]string {
	if existing == nil {
		return patch
	}
	if patch == nil {
		return existing
	}
	for k, v := range patch {
		existing[k] = v
	}
	return existing
}

// mergeOwners merges new owners into existing ones, deduplicating by email or uuid.
func mergeOwners(existing, patch []user.User) []user.User {
	if existing == nil {
		return patch
	}
	if patch == nil {
		return existing
	}

	seen := make(map[string]int) // key -> index in result
	result := make([]user.User, len(existing))
	copy(result, existing)

	for i, o := range result {
		if key := ownerKey(o); key != "" {
			seen[key] = i
		}
	}

	for _, o := range patch {
		key := ownerKey(o)
		if key == "" {
			result = append(result, o)
			continue
		}
		if idx, exists := seen[key]; exists {
			// update existing owner in-place
			result[idx] = o
		} else {
			seen[key] = len(result)
			result = append(result, o)
		}
	}

	return result
}

// ownerKey returns a deduplication key for an owner (email preferred, then uuid).
func ownerKey(o user.User) string {
	if o.Email != "" {
		return "email:" + o.Email
	}
	if o.UUID != "" {
		return "uuid:" + o.UUID
	}
	return ""
}

// buildLabels builds labels from interface{}
func buildLabels(data interface{}) (labels map[string]string) {
	switch d := data.(type) {
	case map[string]interface{}:
		labels = map[string]string{}
		for key, value := range d {
			stringVal, ok := value.(string)
			if !ok {
				continue
			}
			labels[key] = stringVal
		}
	case map[string]string:
		labels = d
	default:
		labels = nil
	}

	return
}

// buildOwners builds owners from interface{}
func buildOwners(data interface{}) (owners []user.User) {
	buildOwner := func(data map[string]interface{}) user.User {
		return user.User{
			ID:       getString("id", data),
			UUID:     getString("uuid", data),
			Email:    getString("email", data),
			Provider: getString("provider", data),
		}
	}

	switch d := data.(type) {
	case []interface{}:
		owners = []user.User{}
		for _, value := range d {
			mapValue, ok := value.(map[string]interface{})
			if !ok {
				continue
			}
			owners = append(owners, buildOwner(mapValue))
		}
	case []map[string]interface{}:
		owners = []user.User{}
		for _, value := range d {
			owners = append(owners, buildOwner(value))
		}
	case []user.User:
		owners = d
	default:
		owners = nil
	}

	return
}

// patchAssetData patches asset's data using map
func patchAssetData(a *Asset, data interface{}) {
	if data == nil {
		return
	}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	if a.Data == nil {
		a.Data = dataMap
		return
	}

	a.Data = mergemap.Merge(a.Data, dataMap)
}

func patchString(key string, data map[string]interface{}, defaultVal string) string {
	_, exists := data[key]
	if !exists {
		return defaultVal
	}

	return getString(key, data)
}

func getString(key string, data map[string]interface{}) string {
	val, exists := data[key]
	if !exists {
		return ""
	}
	stringVal, ok := val.(string)
	if !ok {
		return ""
	}

	return stringVal
}
