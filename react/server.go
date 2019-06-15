package react

import (
	"cynex/conf"
	"cynex/log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var Server *server

type server struct {
	Config       map[string]string // 配置项
	DownloadDir  string            // 文件下载路径
	HTTPPort     int               // HTTP端口
	HTTPSPort    int               // HTTPS端口
	HTTPSEnabled bool              // 是否启用HTTPS
	TLSKeyDir    string            // HTTPS私钥文件
	TLSCertDir   string            // HTTPS证书文件

	handler *handler
}

const (
	defaultHttpPort     string = "80"
	defaultHttpsPort    string = "443"
	defaultDownloadDir  string = "."
	defaultLogThreshold string = "INFO"
)

const (
	// 配置文件路径
	wd      string = ""
	confDir string = "/conf"
	// 配置文件标识
	httpPort      string = "http.port"
	httpsEnable   string = "https.enable"
	httpsPort     string = "https.port"
	httpsKeyFile  string = "https.key_dir"
	httpsCertFile string = "https.cert_dir"
	//
	downloadDir  string = "default.download_dir"
	logDir       string = "default.log_dir"
	logThreshold string = "default.log_threshold"
)

func init() {

	Server = &server{
		handler:     defaultHandler,
		DownloadDir: defaultDownloadDir,
	}
	// 配置加载
	configs := make(map[string]string)
	configsWd, err := conf.Load(wd)
	if err != nil {
		log.Error("配置加载错误：===>", err)
	}
	for key, val := range configsWd {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(val) != "" {
			configs[key] = val
		}
	}
	configsDir, err := conf.Load(confDir)
	if err != nil {
		log.Error("配置加载错误：===>", err)
	}
	for key, val := range configsDir {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(val) != "" {
			configs[key] = val
		}
	}
	Server.Config = configs
	Server.DownloadDir = Server.getDownloadDir(configs)
	Server.HTTPPort, _ = strconv.Atoi(Server.getHttpPort(configs))
	Server.HTTPSPort, _ = strconv.Atoi(Server.getHttpsPort(configs))
	Server.TLSKeyDir = Server.getHttpsKey(configs)
	Server.TLSCertDir = Server.getHttpsCert(configs)
	Server.HTTPSEnabled = Server.getHttpsEnable(configs)

	log.Dir = Server.getLogDir(configs)
	log.Threshold = Server.getLogThreshold(configs)
	log.UseSetting()
}

func (s *server) Start() {

	// 启动服务
	log.Info("正在启动服务...")
	if s.getHttpsEnable(Server.Config) {
		log.Info("启动成功，正在监听:" + strconv.Itoa(s.HTTPSPort) + "端口（HTTPS）")
		go http.ListenAndServeTLS(":"+strconv.Itoa(s.HTTPSPort), s.TLSCertDir, s.TLSKeyDir, s.handler)
	}
	log.Info("启动成功，正在监听:" + strconv.Itoa(s.HTTPPort) + "端口（HTTP）")
	http.ListenAndServe(":"+strconv.Itoa(s.HTTPPort), s.handler)
}

// ConfigValue 读取配置值
func (s *server) ConfigValue(name string) string {
	return s.Config[name]
}

func (s *server) getHttpPort(configs map[string]string) string {
	if configs[httpPort] != "" {
		return configs[httpPort]
	}
	return defaultHttpPort
}

func (s *server) getHttpsEnable(configs map[string]string) bool {
	if configs[httpsEnable] == "" {
		return false
	}
	reg, _ := regexp.Compile("[0-9]")
	if reg.MatchString(configs[httpsEnable]) {
		n, err := strconv.Atoi(configs[httpsEnable])
		if err != nil {
			log.Warning("配置项：https.enable 不符合语法规范")
			return false
		}
		if n > 0 {
			return true
		}
		return false
	}
	var ons = []string{"true", "on", "open", "start"}
	for _, on := range ons {
		if on == strings.ToLower(configs[httpsEnable]) {
			return true
		}
	}
	return false
}

func (s *server) getHttpsPort(configs map[string]string) string {
	if configs[httpsPort] != "" {
		return configs[httpsPort]
	}
	return defaultHttpsPort
}

func (s *server) getHttpsCert(configs map[string]string) string {
	dir := configs[httpsCertFile]
	if dir != "" && strings.Index(dir, "/") != 0 {
		if strings.Index(dir, "./") != 0 {
			wd, _ := os.Getwd()
			dir = wd + "/" + dir
		}
	}
	return dir
}

func (s *server) getHttpsKey(configs map[string]string) string {
	dir := configs[httpsKeyFile]
	if dir != "" && strings.Index(dir, "/") != 0 {
		if strings.Index(dir, "./") != 0 {
			wd, _ := os.Getwd()
			dir = wd + "/" + dir
		}
	}
	return dir
}

func (s *server) getDownloadDir(configs map[string]string) string {
	downDir := configs[downloadDir]
	if downDir != "" && strings.Index(downDir, "/") != 0 {
		if strings.Index(downDir, "./") != 0 {
			wd, _ := os.Getwd()
			downDir = wd + "/" + downDir
		}
	}
	return downDir
}

func (s *server) getLogDir(configs map[string]string) string {
	logDir := configs[logDir]
	if logDir != "" && strings.Index(logDir, "/") != 0 {
		if strings.Index(logDir, "./") != 0 {
			wd, _ := os.Getwd()
			logDir = wd + "/" + logDir
		}
	}
	return logDir
}

func (s *server) getLogThreshold(configs map[string]string) string {
	if configs[logThreshold] != "" {
		return configs[logThreshold]
	}
	return defaultLogThreshold
}
