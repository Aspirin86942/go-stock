package data

// DefaultSponsorAESKeyHex 与旧版赞助码解密逻辑兼容保留。
const DefaultSponsorAESKeyHex = ""

// SponsorDecryptKeyHex 由主程序启动时同步，当前版本仅兼容保留该变量。
var SponsorDecryptKeyHex string

// EffectiveSponsorVipLevel 兼容保留旧接口结构。
// 当前版本已开放全部功能，因此统一返回可用状态。
func EffectiveSponsorVipLevel() (level int, active bool) {
	return 2, true
}
