package encoding

import "time"

// 支持的编码格式
const (
	EncodingUTF8        = "UTF-8"
	EncodingUTF16       = "UTF-16"
	EncodingUTF16LE     = "UTF-16LE"
	EncodingUTF16BE     = "UTF-16BE"
	EncodingUTF32       = "UTF-32"
	EncodingUTF32LE     = "UTF-32LE"
	EncodingUTF32BE     = "UTF-32BE"
	EncodingGBK         = "GBK"
	EncodingGB2312      = "GB2312"
	EncodingGB18030     = "GB18030"
	EncodingBIG5        = "BIG5"
	EncodingShiftJIS    = "SHIFT_JIS"
	EncodingEUCJP       = "EUC-JP"
	EncodingEUCKR       = "EUC-KR"
	EncodingISO88591    = "ISO-8859-1"
	EncodingISO88592    = "ISO-8859-2"
	EncodingISO88595    = "ISO-8859-5"
	EncodingISO885915   = "ISO-8859-15"
	EncodingWindows1250 = "WINDOWS-1250"
	EncodingWindows1251 = "WINDOWS-1251"
	EncodingWindows1252 = "WINDOWS-1252"
	EncodingWindows1254 = "WINDOWS-1254"
	EncodingKOI8R       = "KOI8-R"
	EncodingCP866       = "CP866"
	EncodingMacintosh   = "MACINTOSH"
)

// 操作类型
const (
	OperationDetect   = "detect"
	OperationConvert  = "convert"
	OperationProcess  = "process"
	OperationValidate = "validate"
)

// 默认配置值
const (
	DefaultSampleSize         = 8192        // 默认检测样本大小
	DefaultMinConfidence      = 0.8         // 默认最小置信度
	DefaultBufferSize         = 8192        // 默认缓冲区大小
	DefaultInvalidChar        = "?"         // 默认无效字符替换
	DefaultBackupSuffix       = ".bak"      // 默认备份后缀
	DefaultChunkSize          = 1024 * 1024 // 默认分块大小 (1MB)
	DefaultMaxFileSize        = 100 << 20   // 默认最大文件大小 (100MB)
	DefaultCacheSize          = 1000        // 默认缓存大小
	DefaultCacheTTL           = time.Hour   // 默认缓存过期时间
)

// 换行符常量
const (
	LineEndingLF   = "\n"   // Unix/Linux 换行符
	LineEndingCRLF = "\r\n" // Windows 换行符
	LineEndingCR   = "\r"   // Classic Mac 换行符
)