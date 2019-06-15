// reactor 提供指定路径的处理器绑定
// 支持浏览器Get与Post方法
package react

import (
	"cynex/cache"
	"cynex/log"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

var defaultRouter *router
var defaultHandler *handler

type reactor struct {
	responseWriter http.ResponseWriter // react.responseWriter
	request        *http.Request       // *react.request

	reactorPool  *sync.Pool
	pathVariable map[string][]string
}

type router struct {
	muStatic   *sync.RWMutex
	muDownload *sync.RWMutex

	pathTree  *pTree       // 路径树
	pathCache *cache.Cache // 路由动态缓存

	statics       map[string]string // 静态文件
	staticCache   *cache.Cache      // 静态文件缓存
	downloads     map[string]string // 下载文件
	downloadCache *cache.Cache      // 下载文件缓存

	beforeCache *cache.Cache
	afterCache  *cache.Cache
}

type handler struct {
	reactorPool *sync.Pool
}

func init() {
	pathTree := &pTree{
		root: &pNode{
			name: "ROOT",
			sub:  *new([]*pNode),
			val:  nil,
		},
	}
	defaultRouter = &router{
		pathTree:      pathTree,
		pathCache:     cache.NewCache(),
		statics:       make(map[string]string),
		staticCache:   cache.NewCache(7 * 7),
		downloads:     make(map[string]string),
		downloadCache: cache.NewCache(7 * 7),
		muStatic:      &sync.RWMutex{},
		muDownload:    &sync.RWMutex{},
		beforeCache:   cache.NewCache(math.MaxInt64),
		afterCache:    cache.NewCache(math.MaxInt64),
	}
	defaultHandler = &handler{
		reactorPool: &sync.Pool{
			New: func() interface{} {
				return new(reactor)
			},
		},
	}
}

// BindGet 提供GET方式的HTTP访问
// 参数 url:将要注册处理的请求路径;comp:使用此组件中的方法处理请求;method:使用（指定组件中的）此方法处理请求;
func BindGet(url string, comp interface{}, method string) {
	url = stripLastSlash(url)
	handle := buildHandleFunc(comp, method)
	defaultRouter.pathRegister(url, handle, "GET")
	defaultRouter.pathRegister(url, handle, "OPTIONS")
	log.Info("已绑定GET方法路径：" + url)
}

// BindPost 提供POST方式的HTTP访问
// 参数 url:将要注册处理的请求路径;comp:使用此组件中的方法处理请求;method:使用（指定组件中的）此方法处理请求;
func BindPost(url string, comp interface{}, method string) {
	url = stripLastSlash(url)
	handle := buildHandleFunc(comp, method)
	defaultRouter.pathRegister(url, handle, "POST")
	defaultRouter.pathRegister(url, handle, "OPTIONS")
	log.Info("已绑定POST方法路径：" + url)
}

// PipelineBefore 用于绑定前置管线
func PipelineBefore(comp interface{}, method string, beforeComp interface{}, beforeMethod string) {
	handle := buildHandleFunc(comp, method)
	beforeHandle := buildHandleFunc(beforeComp, beforeMethod)
	defaultRouter.beforeCache.Set(handle.handleFuncDes, beforeHandle)
}

// PipelineAfter 用于绑定后置管线
func PipelineAfter(comp interface{}, method string, afterComp interface{}, afterMethod string) {
	handle := buildHandleFunc(comp, method)
	afterHandle := buildHandleFunc(afterComp, afterMethod)
	defaultRouter.afterCache.Set(handle.handleFuncDes, afterHandle)
}

func buildHandleFunc(comp interface{}, method string) *handleMethodRefer {
	v := reflect.ValueOf(comp)
	handleFunc := v.MethodByName(method)
	handleFuncDes := v.Type().String() + "." + method
	return &handleMethodRefer{
		handleFunc:    handleFunc,
		handleFuncDes: handleFuncDes,
	}
}

// BindStatic 提供静态文件绑定
// 参数 urlPrefix: 绑定路径前缀；localPrefix: 本地路径（工作目录内）前缀
func BindStatic(urlPrefix string, localPrefix string) {
	urlPrefix = stripLastSlash(urlPrefix)
	localPrefix = stripLastSlash(localPrefix)
	localPrefix = completeFirstSlash(localPrefix)
	defaultRouter.muStatic.Lock()
	defaultRouter.statics[urlPrefix] = localPrefix
	defaultRouter.muStatic.Unlock()
	log.Info("已绑定静态文件目录前置路径：" + urlPrefix)
}

// BindDownload 提供文件下载绑定
// 参数 url: 文件请求路径；path: 文件下载路径（将被连接至downloadDir后形成全路径）
func BindDownload(url string, path string) {
	url = stripLastSlash(url)
	path = completeFirstSlash(path)
	defaultRouter.muDownload.Lock()
	defaultRouter.downloads[url] = path
	defaultRouter.muDownload.Unlock()
	log.Info("已绑定下载文件请求路径：" + url)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	re := h.reactorPool.Get().(*reactor)
	re.responseWriter = w
	re.request = r
	re.reactorPool = h.reactorPool
	re.pathVariable = make(map[string][]string)
	h.acceptAndProcess(re)
}

func (h *handler) acceptAndProcess(re *reactor) {
	defer func(re *reactor) {
		re.responseWriter = nil
		re.request = nil
		re.pathVariable = nil
		re.reactorPool.Put(re)
	}(re)
	re.request.ParseForm()
	var uri = re.request.RequestURI
	if qi := strings.Index(uri, "?"); qi > 0 {
		uri = uri[:qi]
	}
	uri = stripLastSlash(uri)
	key := strings.ToUpper(re.request.Method) + ":" + uri
	log.Debug("接收并处理请求===> " + key)
	if val, err := defaultRouter.pathCache.Get(key); err == nil {
		c := val.(*cachedHandleMethodRefer)
		for key, val := range c.pathVars {
			if strings.TrimSpace(key) != "" {
				re.request.Form[key] = val
			}
		}
		c.methodRefer.Call(re.responseWriter, re.request)
	} else {
		if exe, err := re.matchHandler(key); err == nil {
			exe.Call(re.responseWriter, re.request)
			cs := &cachedHandleMethodRefer{
				methodRefer: exe,
				pathVars:    re.pathVariable,
			}
			go defaultRouter.pathCache.Set(key, cs)
		} else {
			// 静态文件处理
			if p, err := re.isStatic(uri); err == nil {
				if err = re.handleStatic(p, re.responseWriter); err != nil {
					re.responseWriter.Write([]byte(err.Error()))
				}
				return
			}
			//./
			// 下载文件处理
			if p, err := re.isDownload(uri); err == nil {
				if err = re.handleDownload(p, re.responseWriter); err != nil {
					re.responseWriter.Write([]byte(err.Error()))
				}
				return
			}
			//./
			re.responseWriter.WriteHeader(http.StatusNotFound)
			re.responseWriter.Write([]byte("Not Found"))
		}
	}
}

// 路径处理注册
func (*router) pathRegister(path string, comp *handleMethodRefer, method string) {
	key := method + ":" + path
	defaultRouter.addPathHandler(key, comp)
}

type pNode struct {
	name string
	val  *handleMethodRefer
	sub  []*pNode
}

type pTree struct {
	root *pNode
}

// 保存路由配置
func (*router) addPathHandler(path string, comp *handleMethodRefer) {
	splits := strings.Split(path, "/")
	cNode := defaultRouter.pathTree.root
	for i, val := range splits {
		if strings.TrimSpace(val) == "" {
			continue
		}
		if i < len(splits)-1 {
			if n, err := defaultRouter.inRouter(cNode.sub, val); err == nil {
				cNode = n
			} else {
				n := new(pNode)
				n.val = nil
				n.name = val
				cNode.sub = append(cNode.sub, n)
				cNode = n
			}
			continue
		} else {
			if n, err := defaultRouter.inRouter(cNode.sub, val); err == nil {
				n.val = comp
			} else {
				n := new(pNode)
				n.val = comp
				n.name = val
				cNode.sub = append(cNode.sub, n)
			}

		}
	}
}

type handleMethodRefer struct {
	handleFunc    reflect.Value // 处理方法
	handleFuncDes string        // 处理方法描述
}

type cachedHandleMethodRefer struct {
	methodRefer *handleMethodRefer
	pathVars    map[string][]string
}

func (tf *handleMethodRefer) Call(w http.ResponseWriter, r *http.Request) {
	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(w)
	in[1] = reflect.ValueOf(r)
	if before, err := defaultRouter.beforeCache.Get(tf.handleFuncDes); err == nil {
		executeBefore(before.(*handleMethodRefer), in)
	}
	tf.handleFunc.Call(in)
	if after, err := defaultRouter.afterCache.Get(tf.handleFuncDes); err == nil {
		executeAfter(after.(*handleMethodRefer), in)
	}
}

func executeBefore(tf *handleMethodRefer, in []reflect.Value) {
	if before, err := defaultRouter.beforeCache.Get(tf.handleFuncDes); err == nil {
		executeBefore(before.(*handleMethodRefer), in)
	}
	tf.handleFunc.Call(in)
}

func executeAfter(tf *handleMethodRefer, in []reflect.Value) {
	tf.handleFunc.Call(in)
	if after, err := defaultRouter.afterCache.Get(tf.handleFuncDes); err == nil {
		executeAfter(after.(*handleMethodRefer), in)
	}
}

// 是否已经设定路由
func (re *router) inRouter(sub []*pNode, name string) (*pNode, error) {
	for _, n := range sub {
		if n.name == name {
			return n, nil
		}
	}
	return nil, errors.New("not exist")
}

// 获取匹配当前请求路径的处理方法
func (re *reactor) matchHandler(path string) (*handleMethodRefer, error) {
	rErr := errors.New("not exist")
	splits := strings.Split(path, "/")
	cNode := defaultRouter.pathTree.root
	for i, val := range splits {
		if strings.TrimSpace(val) == "" {
			continue
		}
		if i < len(splits)-1 {
			if n, err := re.pathMatch(cNode.sub, val); err == nil {
				cNode = n
				continue
			}
			return nil, rErr
		} else {
			if n, err := re.pathMatch(cNode.sub, val); err == nil {
				if n.val == nil {
					return nil, rErr
				}
				return n.val, nil
			}
			return nil, rErr
		}
	}
	return nil, rErr
}

// 是否存在已经保存的路由设置
// 支持全路径匹配、正则匹配、常用匹配、变量路径
func (re *reactor) pathMatch(sub []*pNode, name string) (*pNode, error) {
	rErr := errors.New("not exist")
	for _, n := range sub {
		if strings.Index(n.name, "(") == 0 && strings.Index(n.name, ")") == len(n.name)-1 {
			// 正则匹配
			reg, err := regexp.Compile(n.name[1 : len(n.name)-1])
			if err != nil {
				panic(err)
			}
			if reg.MatchString(name) {
				return n, nil
			}
			return nil, rErr
		}
		if strings.Index(n.name, "[") == 0 && strings.Index(n.name, "]") == len(n.name)-1 {
			// 常用匹配，使用*号匹配常用字符串
			t := n.name[1 : len(n.name)-1]
			if t == "*" {
				return n, nil
			}
			if strings.Index(t, "*") == 0 && strings.LastIndex(t, "*") == len(t)-1 {
				if strings.Contains(name, t[1:len(t)-1]) {
					return n, nil
				}
				return nil, rErr
			}
			if strings.Index(t, "*") == 0 {
				m := t[1:]
				if i := strings.Index(name, m); i > -1 && i+len(m) == len(name) {
					return n, nil
				}
				return nil, rErr
			}
			if strings.Index(t, "*") == len(t)-1 {
				m := t[:len(t)-1]
				if strings.Index(name, m) == 0 {
					return n, nil
				}
				return nil, rErr
			}
		}
		if strings.Index(n.name, "{") == 0 && strings.Index(n.name, "}") == len(n.name)-1 {
			// 变量路径匹配，含有路径变量，将路径中的值作为匹配变量的值存储Request
			formName := n.name[1 : len(n.name)-1]
			formValue := re.request.Form[formName]
			nameBackup := name
			name, err := url.PathUnescape(name)
			if err != nil {
				name = nameBackup
			}
			formValue = append(formValue, name)
			re.request.Form[formName] = formValue
			re.pathVariable[formName] = formValue
			return n, nil
		}
		// 全路径匹配
		if n.name == name {
			return n, nil
		}
	}
	return nil, rErr
}

// 是否是静态文件
func (re *reactor) isStatic(uri string) (string, error) {
	defaultRouter.muStatic.RLock()
	val, err := defaultRouter.staticCache.Get(uri)
	defaultRouter.muStatic.RUnlock()
	if err == nil {
		return val.(string), nil
	}
	for key, val := range defaultRouter.statics {
		if strings.Index(uri, key) == 0 {
			rVal := val + uri[len(key):]
			go func() {
				defaultRouter.muStatic.Lock()
				defaultRouter.staticCache.Set(uri, rVal)
				defaultRouter.muStatic.Unlock()
			}()
			log.Debug("静态文件：" + uri)
			return rVal, nil
		}
	}
	return "", errors.New("not static file")
}

// 处理静态文件
func (re *reactor) handleStatic(filePath string, w http.ResponseWriter) error {
	wd, _ := os.Getwd()
	f, err := os.Open(wd + filePath)
	if err != nil {
		log.Error("静态文件读取错误：" + err.Error())
		return errors.New("no such file or directory")
	}
	if fi, _ := f.Stat(); fi.IsDir() {
		return errors.New("current path is directory")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Copy(w, f)
	return nil
}

// 是否是下载文件
func (re *reactor) isDownload(uri string) (string, error) {
	defaultRouter.muDownload.RLock()
	val, err := defaultRouter.downloadCache.Get(uri)
	defaultRouter.muDownload.RUnlock()
	if err == nil {
		return val.(string), nil
	}
	for key, val := range defaultRouter.downloads {
		if key == uri {
			var pp string
			if Server.DownloadDir == "." || Server.DownloadDir == "./" {
				pp, _ = os.Getwd()
			} else {
				pp = Server.DownloadDir
			}
			rVal := pp + val
			go func() {
				defaultRouter.muDownload.Lock()
				defaultRouter.downloadCache.Set(uri, rVal)
				defaultRouter.muDownload.Unlock()
			}()
			log.Debug("下载文件：" + uri)
			return rVal, nil
		}
	}
	return "", errors.New("not download file")
}

// 处理下载文件
func (re *reactor) handleDownload(filePath string, w http.ResponseWriter) error {
	f, err := os.Open(filePath)
	if err != nil {
		log.Error("下载文件读取错误：" + err.Error())
		return errors.New("no such file or directory")
	}
	fi, _ := f.Stat()
	if fi.IsDir() {
		return errors.New("current path is directory")
	}
	w.Header().Set("Content-Type", "application/octet-stream; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fi.Name()+"\"")
	io.Copy(w, f)
	return nil
}

func stripLastSlash(uri string) string {
	if strings.LastIndex(uri, "/") == len(uri)-1 {
		uri = uri[:len(uri)-1]
	}
	return uri
}

func completeFirstSlash(path string) string {
	if strings.Index(path, "/") != 0 {
		path = "/" + path
	}
	return path
}
