package models

import (
    "bytes"
    "compress/gzip"
    "crypto/tls"
    "fmt"
    "io"
    "log"
    "math/rand"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/PuerkitoBio/goquery"
    "github.com/joho/godotenv"
)

var (
    Client         *http.Client
    sessionCookies []*http.Cookie
    browserProfile *BrowserProfile
    sessionStarted time.Time
)

// BrowserProfile содержит полный профиль браузера
type BrowserProfile struct {
    UserAgent      string
    AcceptLanguage string
    Platform       string
    ScreenSize     string
    TimeZone       string
    WebGLVendor    string
    WebGLRenderer  string
    Plugins        []string
    Fonts          []string
}

// Создаем реалистичные браузерные профили
var browserProfiles = []*BrowserProfile{
    {
        UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        AcceptLanguage: "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
        Platform:       "Win32",
        ScreenSize:     "1920x1080",
        TimeZone:       "Europe/Moscow",
        WebGLVendor:    "Google Inc. (Intel)",
        WebGLRenderer:  "ANGLE (Intel, Intel(R) UHD Graphics 620 Direct3D11 vs_5_0 ps_5_0, D3D11)",
    },
    {
        UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        AcceptLanguage: "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7",
        Platform:       "MacIntel",
        ScreenSize:     "1440x900",
        TimeZone:       "Europe/Moscow",
        WebGLVendor:    "Intel Inc.",
        WebGLRenderer:  "Intel Iris Pro OpenGL Engine",
    },
    {
        UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
        AcceptLanguage: "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3",
        Platform:       "Win32",
        ScreenSize:     "1366x768",
        TimeZone:       "Europe/Moscow",
        WebGLVendor:    "Mozilla",
        WebGLRenderer:  "Mozilla -- ANGLE (Intel, Intel(R) HD Graphics 4000 Direct3D11 vs_5_0 ps_5_0, D3D11)",
    },
}

func InitDB() {
    log.Printf("Starting advanced InitDB...")
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Выбираем случайный профиль браузера
    browserProfile = browserProfiles[rand.Intn(len(browserProfiles))]
    log.Printf("Selected browser profile: %s", strings.Split(browserProfile.UserAgent, " ")[0])

    // Создаем продвинутый HTTP транспорт
    transport := &http.Transport{
        MaxIdleConns:        10,
        MaxIdleConnsPerHost: 2,
        IdleConnTimeout:     60 * time.Second,
        DisableCompression:  false,
        // Имитируем настройки браузера
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
            MaxVersion: tls.VersionTLS13,
            CipherSuites: []uint16{
                tls.TLS_AES_128_GCM_SHA256,
                tls.TLS_AES_256_GCM_SHA384,
                tls.TLS_CHACHA20_POLY1305_SHA256,
                tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
                tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
                tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
                tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            },
        },
    }

    // Настройка прокси
    useProxy := os.Getenv("USE_PROXY") == "true"
    if useProxy {
        proxyURL := os.Getenv("PROXY_URL")
        if proxyURL != "" {
            proxy, err := url.Parse(proxyURL)
            if err != nil {
                log.Fatalf("Invalid PROXY_URL: %v", err)
            }
            transport.Proxy = http.ProxyURL(proxy)
            log.Printf("Proxy configured: %s", proxyURL)
        }
    }

    jar, _ := cookiejar.New(nil)
    Client = &http.Client{
        Transport: transport,
        Jar:       jar,
        Timeout:   60 * time.Second,
    }

    // Многоступенчатая инициализация сессии
    initAdvancedSession()
    sessionStarted = time.Now()
    log.Printf("Advanced HTTP Client initialized")
}

func initAdvancedSession() error {
    log.Printf("Initializing advanced session with multi-step approach...")

    // Шаг 1: Заходим на главную страницу
    if err := visitMainPage(); err != nil {
        return err
    }

    // Шаг 2: Имитируем активность пользователя
    if err := simulateUserActivity(); err != nil {
        log.Printf("Warning: User activity simulation failed: %v", err)
    }

    // Шаг 3: Получаем дополнительные куки
    if err := loadAdditionalResources(); err != nil {
        log.Printf("Warning: Additional resources loading failed: %v", err)
    }

    log.Printf("Advanced session initialized with %d cookies", len(sessionCookies))
    return nil
}

func visitMainPage() error {
    log.Printf("Step 1: Visiting main page...")

    req, err := http.NewRequest("GET", "https://yandex.ru/", nil)
    if err != nil {
        return err
    }

    setAdvancedBrowserHeaders(req, "")

    resp, err := Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    sessionCookies = append(sessionCookies, resp.Cookies()...)

    // Читаем содержимое чтобы показать активность
    io.ReadAll(resp.Body)

    time.Sleep(time.Duration(rand.Intn(3000)+2000) * time.Millisecond)
    return nil
}

func simulateUserActivity() error {
    log.Printf("Step 2: Simulating user activity...")

    // Переходим на страницу поиска
    searchPageURL := "https://yandex.ru/search/"
    req, err := http.NewRequest("GET", searchPageURL, nil)
    if err != nil {
        return err
    }

    setAdvancedBrowserHeaders(req, "https://yandex.ru/")
    addSessionCookies(req)

    resp, err := Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    sessionCookies = append(sessionCookies, resp.Cookies()...)
    io.ReadAll(resp.Body)

    time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
    return nil
}

func loadAdditionalResources() error {
    log.Printf("Step 3: Loading additional resources...")

    // Загружаем некоторые статические ресурсы (имитируем браузер)
    resources := []string{
        "https://yastatic.net/s3/web4lib/_/La6qi18Z8LwgnZdsAr1qy2E.woff2",
        "https://mc.yandex.ru/metrika/tag.js",
    }

    for _, resource := range resources {
        req, err := http.NewRequest("GET", resource, nil)
        if err != nil {
            continue
        }

        setResourceHeaders(req)
        addSessionCookies(req)

        resp, err := Client.Do(req)
        if err != nil {
            continue
        }
        resp.Body.Close()

        time.Sleep(time.Duration(rand.Intn(500)+200) * time.Millisecond)
    }

    return nil
}

func FetchYandexResults(query string, lr string) ([]string, error) {
    log.Printf("Starting advanced FetchYandexResults for query='%s', lr=%s", query, lr)

    // Проверяем возраст сессии
    if time.Since(sessionStarted) > 30*time.Minute {
        log.Printf("Session is old, refreshing...")
        RefreshSession()
    }

    // Предварительная активность
    if err := performPreSearchActivity(query); err != nil {
        log.Printf("Warning: Pre-search activity failed: %v", err)
    }

    // Основной поисковый запрос
    return executeSearchQuery(query, lr)
}

func performPreSearchActivity(query string) error {
    log.Printf("Performing pre-search activity...")

    // Имитируем набор запроса (частичные запросы)
    if len(query) > 3 {
        partialQuery := query[:len(query)/2]
        suggestURL := fmt.Sprintf("https://suggest.yandex.ru/suggest-ya.cgi?part=%s", url.QueryEscape(partialQuery))

        req, err := http.NewRequest("GET", suggestURL, nil)
        if err == nil {
            setAdvancedBrowserHeaders(req, "https://yandex.ru/search/")
            addSessionCookies(req)

            resp, err := Client.Do(req)
            if err == nil {
                resp.Body.Close()
                time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)
            }
        }
    }

    return nil
}

func executeSearchQuery(query, lr string) ([]string, error) {
    // Человекоподобная задержка перед поиском
    time.Sleep(time.Duration(rand.Intn(3000)+2000) * time.Millisecond)

    encodedQuery := url.QueryEscape(query)
    searchURL := fmt.Sprintf("https://yandex.ru/search/?text=%s&lr=%s", encodedQuery, lr)

    // Добавляем случайные параметры как настоящий браузер
    extraParams := []string{
        "&rdrnd=" + strconv.FormatInt(time.Now().UnixNano(), 10),
        "&redircnt=1",
    }
    searchURL += extraParams[rand.Intn(len(extraParams))]

    log.Printf("Constructed search URL: %s", searchURL)

    req, err := http.NewRequest("GET", searchURL, nil)
    if err != nil {
        return nil, err
    }

    setAdvancedBrowserHeaders(req, "https://yandex.ru/search/")
    addSessionCookies(req)

    // Добавляем специфичные для поиска заголовки
    req.Header.Set("Sec-Fetch-Dest", "document")
    req.Header.Set("Sec-Fetch-Mode", "navigate")
    req.Header.Set("Sec-Fetch-Site", "same-origin")
    req.Header.Set("Sec-Fetch-User", "?1")

    resp, err := Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Обновляем куки
    newCookies := resp.Cookies()
    sessionCookies = append(sessionCookies, newCookies...)

    if resp.StatusCode != http.StatusOK {
        return handleAdvancedErrorResponse(resp, encodedQuery, lr)
    }

    return parseAdvancedSearchResults(resp)
}

func setAdvancedBrowserHeaders(req *http.Request, referer string) {
    req.Header.Set("User-Agent", browserProfile.UserAgent)
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
    req.Header.Set("Accept-Language", browserProfile.AcceptLanguage)
    req.Header.Set("Accept-Encoding", "gzip, deflate, br")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Upgrade-Insecure-Requests", "1")
    req.Header.Set("Cache-Control", "max-age=0")

    if referer != "" {
        req.Header.Set("Referer", referer)
    }

    // Браузер-специфичные заголовки
    if strings.Contains(browserProfile.UserAgent, "Chrome") {
        req.Header.Set("sec-ch-ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
        req.Header.Set("sec-ch-ua-mobile", "?0")
        req.Header.Set("sec-ch-ua-platform", fmt.Sprintf(`"%s"`, browserProfile.Platform))
    }
}

func setResourceHeaders(req *http.Request) {
    req.Header.Set("User-Agent", browserProfile.UserAgent)
    req.Header.Set("Accept", "*/*")
    req.Header.Set("Accept-Language", browserProfile.AcceptLanguage)
    req.Header.Set("Accept-Encoding", "gzip, deflate, br")
    req.Header.Set("Connection", "keep-alive")
    req.Header.Set("Sec-Fetch-Dest", "font")
    req.Header.Set("Sec-Fetch-Mode", "cors")
    req.Header.Set("Sec-Fetch-Site", "cross-site")
}

func addSessionCookies(req *http.Request) {
    for _, cookie := range sessionCookies {
        if cookie.Domain == "" || strings.HasSuffix(req.URL.Host, cookie.Domain) {
            req.AddCookie(cookie)
        }
    }
}

func handleAdvancedErrorResponse(resp *http.Response, query, lr string) ([]string, error) {
    bodyBytes, _ := io.ReadAll(resp.Body)
    htmlContent := string(bodyBytes)

    // Сохраняем для анализа
    saveHTMLResponse(bodyBytes, query, lr, fmt.Sprintf("error_%d", resp.StatusCode))

    // Детальный анализ ошибки
    if strings.Contains(htmlContent, "captcha") || strings.Contains(htmlContent, "Введите символы") {
        log.Printf("CAPTCHA detected - need to solve or change approach")
        return nil, fmt.Errorf("CAPTCHA required - session compromised")
    }

    if strings.Contains(htmlContent, "robot") || strings.Contains(htmlContent, "бот") {
        log.Printf("Bot detection triggered")
        return nil, fmt.Errorf("bot detection - need new session and IP")
    }

    if resp.StatusCode == 429 {
        log.Printf("Rate limit exceeded")
        return nil, fmt.Errorf("rate limit exceeded - need longer delays")
    }

    if resp.StatusCode >= 500 {
        log.Printf("Server error: %d", resp.StatusCode)
        return nil, fmt.Errorf("server error: %d", resp.StatusCode)
    }

    return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
}

func parseAdvancedSearchResults(resp *http.Response) ([]string, error) {
    var bodyReader io.Reader = resp.Body
    if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
        reader, err := gzip.NewReader(resp.Body)
        if err != nil {
            return nil, err
        }
        defer reader.Close()
        bodyReader = reader
    }

    bodyBytes, err := io.ReadAll(bodyReader)
    if err != nil {
        return nil, err
    }

    // Сохраняем для анализа
    saveHTMLResponse(bodyBytes, "search", "success", "success")

    doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
    if err != nil {
        return nil, err
    }

    var links []string

    // Расширенные селекторы для Яндекса
    selectors := []string{
        "li.serp-item .OrganicTitle-Link",
        "li.serp-item h2 a",
        ".serp-item .organic__url a",
        ".serp-item a[href*='://']:not([href*='yandex.ru'])",
        ".Organic .OrganicTitle-Link",
        ".VanillaReact .OrganicTitle-Link",
        "[data-cid] a[href^='httpRequest']:not([href*='yandex.ru/clck'])",
    }

    for _, selector := range selectors {
        found := 0
        doc.Find(selector).Each(func(i int, s *goquery.Selection) {
            if href, exists := s.Attr("href"); exists {
                cleanURL := extractCleanURL(href)
                if cleanURL != "" && isValidSearchResult(cleanURL) {
                    links = append(links, cleanURL)
                    found++
                }
            }
        })

        log.Printf("Selector '%s' found %d links", selector, found)
        if len(links) >= 10 { // Достаточно результатов
            break
        }
    }

    // Удаляем дубликаты
    links = removeDuplicates(links)

    log.Printf("Extracted %d unique links from response", len(links))
    return links, nil
}

func extractCleanURL(href string) string {
    // Обработка Яндекс-редиректов
    if strings.Contains(href, "yandex.ru/clck/") {
        re := regexp.MustCompile(`text=([^&]+)`)
        matches := re.FindStringSubmatch(href)
        if len(matches) > 1 {
            if decodedURL, err := url.QueryUnescape(matches[1]); err == nil {
                return decodedURL
            }
        }
    }

    // Прямые ссылки
    if strings.HasPrefix(href, "httpRequest") && !strings.Contains(href, "yandex.ru") {
        return href
    }

    return ""
}

func isValidSearchResult(url string) bool {
    // Фильтруем нежелательные результаты
    excludePatterns := []string{
        "yandex.ru",
        "ya.ru",
        "yandex.net",
        "yastatic.net",
        "javascript:",
        "mailto:",
        "tel:",
    }

    for _, pattern := range excludePatterns {
        if strings.Contains(url, pattern) {
            return false
        }
    }

    return len(url) > 10 && (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
}

func RefreshSession() error {
    log.Printf("Refreshing session with full cleanup...")

    // Очищаем все данные сессии
    sessionCookies = nil

    // Меняем профиль браузера
    browserProfile = browserProfiles[rand.Intn(len(browserProfiles))]
    log.Printf("Changed to browser profile: %s", strings.Split(browserProfile.UserAgent, " ")[0])

    // Переинициализируем сессию
    if err := initAdvancedSession(); err != nil {
        return err
    }

    sessionStarted = time.Now()
    return nil
}

func WaitBetweenRequests() {
    // Увеличенные задержки для большей натуральности
    delay := time.Duration(rand.Intn(45000)+30000) * time.Millisecond // 30-75 секунд
    log.Printf("Waiting %v before next request (human-like behavior)", delay)
    time.Sleep(delay)
}

func saveHTMLResponse(bodyBytes []byte, query, lr, suffix string) {
    htmlFileName := fmt.Sprintf("debug_%s_%s_%d_%s.html", query, lr, time.Now().Unix(), suffix)
    htmlFile, err := os.Create(htmlFileName)
    if err != nil {
        log.Printf("Error creating HTML file %s: %v", htmlFileName, err)
        return
    }
    defer htmlFile.Close()

    if _, err := htmlFile.Write(bodyBytes); err != nil {
        log.Printf("Error writing to HTML file %s: %v", htmlFileName, err)
        return
    }
    log.Printf("HTML saved to file: %s", htmlFileName)
}

func removeDuplicates(links []string) []string {
    seen := make(map[string]bool)
    var result []string

    for _, link := range links {
        if !seen[link] {
            seen[link] = true
            result = append(result, link)
        }
    }

    return result
}
