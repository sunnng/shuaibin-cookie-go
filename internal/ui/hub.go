package ui

// Task Hub / 导航相关 Store key（与 binding 四件套中的业务 key 区分）。
const (
	KeyHubNav      = "hub_nav"      // modules | license | system
	KeyHubCat      = "hub_cat"      // all | daily | event | maint
	KeyHubSelected = "hub_selected" // module id
	KeyHubFilter   = "hub_filter"   // 搜索关键字

	HubNavModules = "modules"
	HubNavLicense = "license"
	HubNavSystem  = "system"

	HubCatAll   = "all"
	HubCatDaily = "daily"
	HubCatEvent = "event"
	HubCatMaint = "maint"

	KeyLicenseStatus = "license_status" // unverified | ok | expired（占位）
)

// SeedHubDefaults 填充 Task Hub 导航默认值（仅缺失 key）。
func SeedHubDefaults(store *Store) {
	if store == nil {
		return
	}
	if !store.HasKey(KeyHubNav) {
		store.SetString(KeyHubNav, HubNavModules)
	}
	if !store.HasKey(KeyHubCat) {
		store.SetString(KeyHubCat, HubCatAll)
	}
	if !store.HasKey(KeyHubSelected) || store.GetString(KeyHubSelected) == "" {
		mods := Modules()
		if len(mods) > 0 {
			store.SetString(KeyHubSelected, mods[0].ID)
		} else {
			store.SetString(KeyHubSelected, "")
		}
	}
	if !store.HasKey(KeyHubFilter) {
		store.SetString(KeyHubFilter, "")
	}
	if !store.HasKey(KeyLicenseStatus) {
		store.SetString(KeyLicenseStatus, "unverified")
	}
}
