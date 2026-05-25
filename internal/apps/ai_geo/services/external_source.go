package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
	"gorm.io/gorm"

	"saas_baseon_go/internal/apps/ai_geo/dto"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

const (
	externalSourceMaxBytes = 2 << 20
	externalSourceMaxText  = 60000
)

type fetchedExternalPage struct {
	URL         string
	Site        string
	Title       string
	Description string
	Text        string
}

func (s *Service) ExtractExternalSource(ctx context.Context, viewer dto.Viewer, payload dto.ExternalSourceExtractPayload) (dto.ExternalSourceExtractResult, error) {
	extractType := strings.ToLower(strings.TrimSpace(payload.ExtractType))
	if extractType == "" {
		extractType = "style"
	}
	if viewer.TenantID == 0 || (extractType != "style" && extractType != "hotspot") {
		return dto.ExternalSourceExtractResult{}, ErrInvalidInput
	}
	page, err := fetchExternalPage(ctx, payload.URL)
	if err != nil {
		return dto.ExternalSourceExtractResult{}, err
	}
	source, err := s.upsertExternalSource(ctx, viewer, page)
	if err != nil {
		return dto.ExternalSourceExtractResult{}, err
	}
	result := dto.ExternalSourceExtractResult{Source: source}
	if extractType == "style" {
		style, err := s.createStyleTemplateFromSource(ctx, viewer, source, payload)
		if err != nil {
			return result, err
		}
		result.StyleTemplate = &style
		return result, nil
	}
	result.HotspotDraft = hotspotDraftFromSource(source, payload)
	return result, nil
}

func (s *Service) StyleTemplates(ctx context.Context, viewer dto.Viewer, req dto.PageRequest) (dto.PageResponse[models.AiGeoStyleTemplate], error) {
	rows, total, err := s.repo.ListStyleTemplates(ctx, viewer.TenantID, req)
	return page(rows, total, req), err
}

func (s *Service) StyleTemplate(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoStyleTemplate, error) {
	row, err := s.repo.StyleTemplate(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	return row, err
}

func (s *Service) CreateStyleTemplate(ctx context.Context, viewer dto.Viewer, payload dto.StyleTemplatePayload) (models.AiGeoStyleTemplate, error) {
	name := strings.TrimSpace(payload.TemplateName)
	if viewer.TenantID == 0 || name == "" {
		return models.AiGeoStyleTemplate{}, ErrInvalidInput
	}
	if payload.SourceID != nil && *payload.SourceID > 0 {
		if _, err := s.repo.ExternalSource(ctx, viewer.TenantID, *payload.SourceID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return models.AiGeoStyleTemplate{}, ErrNotFound
			}
			return models.AiGeoStyleTemplate{}, err
		}
	}
	now := time.Now()
	userID := viewer.UserID
	row := models.AiGeoStyleTemplate{
		TenantID:          viewer.TenantID,
		SourceID:          payload.SourceID,
		TemplateCode:      defaultString(strings.TrimSpace(payload.TemplateCode), fmt.Sprintf("style-%d", now.UnixNano())),
		TemplateName:      name,
		Description:       stringPtr(payload.Description),
		ContentType:       stringPtr(payload.ContentType),
		Platform:          stringPtr(payload.Platform),
		ToneProfile:       jsonString(payload.ToneProfile, map[string]interface{}{}),
		StructureProfile:  jsonString(payload.StructureProfile, map[string]interface{}{}),
		TechniqueProfile:  jsonString(payload.TechniqueProfile, map[string]interface{}{}),
		StyleKeywords:     jsonString(payload.StyleKeywords, []string{}),
		PromptFragment:    strings.TrimSpace(payload.PromptFragment),
		NegativeRules:     jsonString(payload.NegativeRules, []string{}),
		ExtractionSummary: jsonString(payload.ExtractionSummary, map[string]interface{}{}),
		Status:            defaultString(payload.Status, "active"),
		CreatedBy:         &userID,
		UpdatedBy:         &userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	err := s.repo.SaveStyleTemplate(ctx, &row)
	return row, err
}

func (s *Service) UpdateStyleTemplate(ctx context.Context, viewer dto.Viewer, id uint64, payload dto.StyleTemplatePayload) (models.AiGeoStyleTemplate, error) {
	row, err := s.repo.StyleTemplate(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	if strings.TrimSpace(payload.TemplateName) != "" {
		row.TemplateName = strings.TrimSpace(payload.TemplateName)
	}
	if strings.TrimSpace(payload.Description) != "" {
		row.Description = stringPtr(payload.Description)
	}
	if strings.TrimSpace(payload.ContentType) != "" {
		row.ContentType = stringPtr(payload.ContentType)
	}
	if strings.TrimSpace(payload.Platform) != "" {
		row.Platform = stringPtr(payload.Platform)
	}
	if payload.ToneProfile != nil {
		row.ToneProfile = jsonString(payload.ToneProfile, map[string]interface{}{})
	}
	if payload.StructureProfile != nil {
		row.StructureProfile = jsonString(payload.StructureProfile, map[string]interface{}{})
	}
	if payload.TechniqueProfile != nil {
		row.TechniqueProfile = jsonString(payload.TechniqueProfile, map[string]interface{}{})
	}
	if payload.StyleKeywords != nil {
		row.StyleKeywords = jsonString(payload.StyleKeywords, []string{})
	}
	if strings.TrimSpace(payload.PromptFragment) != "" {
		row.PromptFragment = strings.TrimSpace(payload.PromptFragment)
	}
	if payload.NegativeRules != nil {
		row.NegativeRules = jsonString(payload.NegativeRules, []string{})
	}
	if payload.ExtractionSummary != nil {
		row.ExtractionSummary = jsonString(payload.ExtractionSummary, map[string]interface{}{})
	}
	if strings.TrimSpace(payload.Status) != "" {
		row.Status = strings.TrimSpace(payload.Status)
	}
	now := time.Now()
	userID := viewer.UserID
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.SaveStyleTemplate(ctx, &row)
	return row, err
}

func (s *Service) ArchiveStyleTemplate(ctx context.Context, viewer dto.Viewer, id uint64) (models.AiGeoStyleTemplate, error) {
	row, err := s.repo.StyleTemplate(ctx, viewer.TenantID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return row, ErrNotFound
	}
	if err != nil {
		return row, err
	}
	now := time.Now()
	userID := viewer.UserID
	row.Status = "archived"
	row.DeletedAt = &now
	row.UpdatedAt = now
	row.UpdatedBy = &userID
	err = s.repo.SaveStyleTemplate(ctx, &row)
	return row, err
}

func (s *Service) upsertExternalSource(ctx context.Context, viewer dto.Viewer, page fetchedExternalPage) (models.AiGeoExternalSource, error) {
	if row, err := s.repo.ExternalSourceByURL(ctx, viewer.TenantID, page.URL); err == nil && strings.TrimSpace(row.CleanText) != "" {
		return row, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.AiGeoExternalSource{}, err
	}
	now := time.Now()
	userID := viewer.UserID
	sum := sha256.Sum256([]byte(page.Text))
	source := models.AiGeoExternalSource{
		TenantID:         viewer.TenantID,
		SourceType:       "url",
		SourceURL:        page.URL,
		SourceSite:       stringPtr(page.Site),
		SourceTitle:      stringPtr(page.Title),
		RawText:          trimText(page.Text, externalSourceMaxText),
		CleanText:        trimText(page.Text, externalSourceMaxText),
		ContentHash:      hex.EncodeToString(sum[:]),
		ExtractedMeta:    jsonString(map[string]interface{}{"description": page.Description}, map[string]interface{}{}),
		ExtractionStatus: "success",
		CapturedAt:       now,
		Status:           "active",
		CreatedBy:        &userID,
		UpdatedBy:        &userID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	err := s.repo.SaveExternalSource(ctx, &source)
	return source, err
}

func (s *Service) createStyleTemplateFromSource(ctx context.Context, viewer dto.Viewer, source models.AiGeoExternalSource, payload dto.ExternalSourceExtractPayload) (models.AiGeoStyleTemplate, error) {
	card := styleCardFromText(source, payload)
	now := time.Now()
	userID := viewer.UserID
	sourceID := source.ID
	row := models.AiGeoStyleTemplate{
		TenantID:          viewer.TenantID,
		SourceID:          &sourceID,
		TemplateCode:      fmt.Sprintf("style-%d-%d", source.ID, now.UnixNano()),
		TemplateName:      strings.TrimSpace(firstNonEmpty(stringValueFromPtr(source.SourceTitle), "参考写作风格")),
		Description:       stringPtr(strings.TrimSpace(payload.Instruction)),
		ContentType:       stringPtr(payload.ContentType),
		Platform:          stringPtr(payload.Platform),
		ToneProfile:       jsonString(card["tone_profile"], map[string]interface{}{}),
		StructureProfile:  jsonString(card["structure_profile"], map[string]interface{}{}),
		TechniqueProfile:  jsonString(card["technique_profile"], map[string]interface{}{}),
		StyleKeywords:     jsonString(card["style_keywords"], []string{}),
		PromptFragment:    fmt.Sprintf("%v", card["prompt_fragment"]),
		NegativeRules:     jsonString(card["negative_rules"], []string{}),
		ExtractionSummary: jsonString(card, map[string]interface{}{}),
		Status:            "active",
		CreatedBy:         &userID,
		UpdatedBy:         &userID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	err := s.repo.SaveStyleTemplate(ctx, &row)
	return row, err
}

func fetchExternalPage(ctx context.Context, rawURL string) (fetchedExternalPage, error) {
	normalized, err := validateExternalURL(rawURL)
	if err != nil {
		return fetchedExternalPage{}, err
	}
	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return errors.New("外部链接重定向次数过多")
			}
			_, err := validateExternalURL(req.URL.String())
			return err
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normalized, nil)
	if err != nil {
		return fetchedExternalPage{}, ErrInvalidInput
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36 AI-GEO-ReferenceFetcher/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	resp, err := client.Do(req)
	if err != nil {
		return fetchedExternalPage{}, fmt.Errorf("外部网页读取失败")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fetchedExternalPage{}, fmt.Errorf("外部网页返回状态异常")
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if contentType != "" && !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		return fetchedExternalPage{}, fmt.Errorf("只支持读取 HTML 网页")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, externalSourceMaxBytes+1))
	if err != nil {
		return fetchedExternalPage{}, fmt.Errorf("外部网页读取失败")
	}
	if len(body) > externalSourceMaxBytes {
		return fetchedExternalPage{}, fmt.Errorf("外部网页内容过大")
	}
	page, err := extractHTMLText(resp.Request.URL.String(), string(body))
	if err != nil {
		return fetchedExternalPage{}, err
	}
	if strings.TrimSpace(page.Text) == "" {
		return fetchedExternalPage{}, fmt.Errorf("未提取到可用正文，请确认链接是公开文章页")
	}
	return page, nil
}

func validateExternalURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidInput
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("只支持 http/https 外部链接")
	}
	host := parsed.Hostname()
	if host == "" || strings.EqualFold(host, "localhost") {
		return "", fmt.Errorf("不允许读取本机或内网地址")
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return "", fmt.Errorf("外部链接域名解析失败")
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return "", fmt.Errorf("不允许读取本机或内网地址")
		}
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func isPublicIP(ip net.IP) bool {
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return false
	}
	return true
}

func extractHTMLText(pageURL, rawHTML string) (fetchedExternalPage, error) {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return fetchedExternalPage{}, fmt.Errorf("网页 HTML 解析失败")
	}
	var title string
	var description string
	var chunks []string
	var candidates []string
	var walk func(*html.Node, bool)
	walk = func(n *html.Node, skip bool) {
		if n.Type == html.ElementNode {
			name := strings.ToLower(n.Data)
			if name == "script" && strings.Contains(strings.ToLower(attrValue(n, "type")), "ld+json") {
				ldTitle, ldText, ldDesc := extractJSONLDArticle(nodeText(n))
				if title == "" {
					title = ldTitle
				}
				if description == "" {
					description = ldDesc
				}
				if len([]rune(ldText)) >= 20 {
					candidates = append(candidates, ldText)
				}
			}
			if name == "script" || name == "style" || name == "noscript" || name == "svg" || name == "nav" || name == "footer" || name == "form" {
				skip = true
			}
			if name == "title" {
				title = strings.TrimSpace(nodeText(n))
			}
			if name == "meta" {
				var metaName, content string
				for _, attr := range n.Attr {
					if strings.EqualFold(attr.Key, "name") || strings.EqualFold(attr.Key, "property") {
						metaName = strings.ToLower(attr.Val)
					}
					if strings.EqualFold(attr.Key, "content") {
						content = attr.Val
					}
				}
				if metaName == "description" || metaName == "og:description" {
					description = strings.TrimSpace(content)
				}
				if title == "" && (metaName == "og:title" || metaName == "twitter:title") {
					title = strings.TrimSpace(content)
				}
			}
			if !skip && (name == "article" || name == "main" || hasArticleHint(n)) {
				text := normalizeArticleText(nodeText(n))
				if len([]rune(text)) >= 20 {
					candidates = append(candidates, text)
				}
			}
			if !skip && (name == "h1" || name == "h2" || name == "h3" || name == "p" || name == "li" || name == "blockquote") {
				text := normalizeSpaces(nodeText(n))
				if len([]rune(text)) >= 8 {
					chunks = append(chunks, text)
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child, skip)
		}
	}
	walk(doc, false)
	if title == "" && len(chunks) > 0 {
		title = trimRunes(chunks[0], 80)
	}
	text := bestArticleText(candidates, chunks, description)
	if isBlockedVerificationPage(title, text, rawHTML) {
		return fetchedExternalPage{}, fmt.Errorf("目标网页需要安全验证，无法自动提取正文")
	}
	parsed, _ := url.Parse(pageURL)
	return fetchedExternalPage{
		URL:         pageURL,
		Site:        parsed.Hostname(),
		Title:       trimRunes(title, 120),
		Description: trimRunes(description, 300),
		Text:        trimText(text, externalSourceMaxText),
	}, nil
}

func attrValue(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val
		}
	}
	return ""
}

func hasArticleHint(n *html.Node) bool {
	value := strings.ToLower(attrValue(n, "id") + " " + attrValue(n, "class") + " " + attrValue(n, "itemprop"))
	return strings.Contains(value, "article") ||
		strings.Contains(value, "content") ||
		strings.Contains(value, "detail") ||
		strings.Contains(value, "rich_text") ||
		strings.Contains(value, "main-text") ||
		strings.Contains(value, "body")
}

func extractJSONLDArticle(raw string) (string, string, string) {
	var data interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &data); err != nil {
		return "", "", ""
	}
	var title, body, desc string
	var scan func(interface{})
	scan = func(v interface{}) {
		switch x := v.(type) {
		case []interface{}:
			for _, item := range x {
				scan(item)
			}
		case map[string]interface{}:
			if title == "" {
				title = stringFromJSON(x["headline"])
			}
			if title == "" {
				title = stringFromJSON(x["name"])
			}
			if body == "" {
				body = stringFromJSON(x["articleBody"])
			}
			if desc == "" {
				desc = stringFromJSON(x["description"])
			}
			for _, item := range x {
				scan(item)
			}
		}
	}
	scan(data)
	return normalizeSpaces(title), normalizeArticleText(body), normalizeSpaces(desc)
}

func stringFromJSON(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if s := stringFromJSON(item); s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, " ")
	default:
		return ""
	}
}

func bestArticleText(candidates []string, chunks []string, description string) string {
	best := ""
	for _, item := range candidates {
		item = normalizeArticleText(item)
		if len([]rune(item)) > len([]rune(best)) {
			best = item
		}
	}
	if best == "" || (len([]rune(best)) < 80 && len(candidates) == 0) {
		best = normalizeArticleText(strings.Join(uniqueStrings(chunks), "\n\n"))
	}
	if best == "" && strings.TrimSpace(description) != "" {
		best = normalizeArticleText(description)
	}
	return best
}

func normalizeArticleText(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = normalizeSpaces(line)
		if len([]rune(line)) >= 2 {
			out = append(out, line)
		}
	}
	return strings.Join(uniqueStrings(out), "\n\n")
}

func isBlockedVerificationPage(title string, text string, rawHTML string) bool {
	combined := title + "\n" + text + "\n" + rawHTML
	blockedMarkers := []string{"百度安全验证", "安全验证", "验证码", "网络不给力，请稍后重试", "verify", "captcha"}
	for _, marker := range blockedMarkers {
		if strings.Contains(strings.ToLower(combined), strings.ToLower(marker)) {
			return len([]rune(text)) < 80
		}
	}
	return false
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
			b.WriteString(" ")
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return b.String()
}

func styleCardFromText(source models.AiGeoExternalSource, payload dto.ExternalSourceExtractPayload) map[string]interface{} {
	text := source.CleanText
	sentences := splitSentences(text)
	avgLen := averageSentenceLength(sentences)
	keywords := styleKeywords(text, payload)
	tone := map[string]interface{}{
		"voice":            toneVoice(text),
		"sentence_density": sentenceDensity(avgLen),
		"point_of_view":    pointOfView(text),
	}
	structure := map[string]interface{}{
		"opening":        openingPattern(sentences),
		"flow":           structureFlow(text),
		"paragraph_note": paragraphNote(text),
	}
	technique := map[string]interface{}{
		"techniques": techniqueList(text),
		"headline":   headlinePattern(stringValueFromPtr(source.SourceTitle)),
	}
	prompt := fmt.Sprintf("参考写作风格：%s；结构采用%s；可迁移手法：%s。只学习语气、结构和表达手法，不复制原文句子，不照搬外部事实。",
		tone["voice"], structure["opening"], strings.Join(asStringSlice(technique["techniques"]), "、"))
	return map[string]interface{}{
		"tone_profile":      tone,
		"structure_profile": structure,
		"technique_profile": technique,
		"style_keywords":    keywords,
		"prompt_fragment":   prompt,
		"negative_rules":    []string{"不得复制原文句子", "不得照搬外部事实", "不得输出与品牌资料冲突的信息"},
		"source_excerpt":    trimRunes(text, 800),
	}
}

func hotspotDraftFromSource(source models.AiGeoExternalSource, payload dto.ExternalSourceExtractPayload) map[string]interface{} {
	text := source.CleanText
	keywords := styleKeywords(text, payload)
	return map[string]interface{}{
		"title":          firstNonEmpty(stringValueFromPtr(source.SourceTitle), "外部文章热点"),
		"platform":       firstNonEmpty(strings.TrimSpace(payload.Platform), stringValueFromPtr(source.SourceSite), "外部来源"),
		"summary":        trimRunes(firstParagraph(text), 260),
		"search_intent":  inferSearchIntent(text),
		"topic_angles":   structureFlow(text),
		"keywords":       keywords,
		"heat_score":     60,
		"risk_notes":     []string{"外部来源需人工确认真实性和时效性", "生成时仅做轻引用"},
		"source_id":      source.ID,
		"source_url":     source.SourceURL,
		"captured_at":    source.CapturedAt,
		"extract_status": "preview",
	}
}

func splitSentences(text string) []string {
	re := regexp.MustCompile(`[。！？!?]\s*`)
	parts := re.Split(text, -1)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = normalizeSpaces(part)
		if len([]rune(part)) >= 6 {
			out = append(out, part)
		}
	}
	return out
}

func averageSentenceLength(sentences []string) int {
	if len(sentences) == 0 {
		return 0
	}
	total := 0
	for _, sentence := range sentences {
		total += len([]rune(sentence))
	}
	return total / len(sentences)
}

func toneVoice(text string) string {
	switch {
	case strings.Contains(text, "你") || strings.Contains(text, "我们"):
		return "对话感强，偏朋友式建议"
	case strings.Contains(text, "为什么") || strings.Contains(text, "如何") || strings.Contains(text, "怎么"):
		return "问题驱动，偏解释型表达"
	default:
		return "信息整理型，语气相对克制"
	}
}

func sentenceDensity(avg int) string {
	switch {
	case avg == 0:
		return "未知"
	case avg <= 28:
		return "短句为主，节奏快"
	case avg <= 55:
		return "中等句长，解释和判断平衡"
	default:
		return "长句较多，偏深度说明"
	}
}

func pointOfView(text string) string {
	if strings.Contains(text, "你") {
		return "第二人称"
	}
	if strings.Contains(text, "我们") {
		return "第一人称复数"
	}
	return "第三人称或客观叙述"
}

func openingPattern(sentences []string) string {
	if len(sentences) == 0 {
		return "先给结论再展开"
	}
	first := sentences[0]
	if strings.ContainsAny(first, "?？") || strings.Contains(first, "为什么") || strings.Contains(first, "怎么") || strings.Contains(first, "如何") {
		return "问题式开场"
	}
	if strings.Contains(first, "如果") || strings.Contains(first, "当你") {
		return "场景代入式开场"
	}
	return "观点陈述式开场"
}

func structureFlow(text string) []string {
	flow := []string{"提出核心判断", "拆解用户场景", "给出选择理由"}
	if strings.Contains(text, "FAQ") || strings.Contains(text, "问") {
		flow = append(flow, "补充问答")
	}
	if strings.Contains(text, "对比") || strings.Contains(text, "区别") || strings.Contains(text, "还是") {
		flow = append(flow, "做差异对比")
	}
	return uniqueStrings(flow)
}

func paragraphNote(text string) string {
	lines := strings.Split(text, "\n")
	short := 0
	for _, line := range lines {
		if l := len([]rune(strings.TrimSpace(line))); l > 0 && l <= 40 {
			short++
		}
	}
	if short >= len(lines)/2 {
		return "短段落较多，适合移动端阅读"
	}
	return "段落承载信息较多，适合深度说明"
}

func techniqueList(text string) []string {
	items := []string{"关键词前置", "结论先行"}
	if strings.Contains(text, "如果") {
		items = append(items, "条件句引导")
	}
	if strings.Contains(text, "对比") || strings.Contains(text, "还是") {
		items = append(items, "二选一对比")
	}
	if strings.Contains(text, "场景") || strings.Contains(text, "通勤") || strings.Contains(text, "日常") {
		items = append(items, "场景化承接")
	}
	return uniqueStrings(items)
}

func headlinePattern(title string) string {
	if strings.ContainsAny(title, "?？") {
		return "问题式标题"
	}
	if strings.Contains(title, "还是") || strings.Contains(title, "VS") || strings.Contains(title, "vs") {
		return "对比式标题"
	}
	return "主题直给式标题"
}

func styleKeywords(text string, payload dto.ExternalSourceExtractPayload) []string {
	candidates := []string{payload.ContentType, payload.Platform}
	for _, token := range []string{"短句", "对比", "场景", "FAQ", "种草", "测评", "趋势", "选择", "甜酷", "工装", "通勤"} {
		if strings.Contains(text, token) {
			candidates = append(candidates, token)
		}
	}
	return uniqueStrings(candidates)
}

func inferSearchIntent(text string) string {
	if strings.Contains(text, "怎么") || strings.Contains(text, "如何") {
		return "方法查询"
	}
	if strings.Contains(text, "还是") || strings.Contains(text, "区别") {
		return "对比决策"
	}
	return "话题了解"
}

func firstParagraph(text string) string {
	for _, part := range strings.Split(text, "\n") {
		part = strings.TrimSpace(part)
		if part != "" {
			return part
		}
	}
	return ""
}

func uniqueStrings(items []string) []string {
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func normalizeSpaces(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func trimText(value string, max int) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= max {
		return value
	}
	return trimRunes(value, max)
}

func asStringSlice(value interface{}) []string {
	items, ok := value.([]string)
	if ok {
		return items
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
