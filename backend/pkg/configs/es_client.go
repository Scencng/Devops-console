package configs

import (
	"context"
	"crypto/tls"
	"devops-console-backend/internal/dal"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
)

var esClientsMutex sync.RWMutex

var EsClients = make(map[uint]*elasticsearch.Client)

// InitEsClients 闂佸憡甯楃换鍌烇綖閹版澘绀?es 闁诲骸绠嶉崹娲春濞戞氨鍗?
func InitEsClients() {
	if GORMDB == nil {
        logs.Warning(nil, "Database not initialized, skip Elasticsearch client initialization")
		return
	}

	// 婵炶揪缍€濞夋洟寮ˉ鐠扲M闂佸搫琚崕鎾敋濡ゅ啠鍋撻崷顓炰粧缂佽翰鍎抽幏鐘绘晜閽樺澹堥柣鐔哥懃濡霉?
	instanceDetails, err := getElasticsearchInstances()
	if err != nil {
        logs.Warning(map[string]interface{}{"error": err.Error()}, "Failed to load Elasticsearch instances, skipping initialization")
		return
	}

	// 婵炶揪缍€濞夋洟寮妶鍥ㄥ闁靛牆鐗忓Σ鍫ユ⒑閺夎法肖闁汇倕妫濆畷姘槈濡偐澶?es 闁诲骸绠嶉崹娲春濞戞氨鍗?
	for _, instanceDetail := range instanceDetails {
		LockEsClients() // 闂佸憡姊绘繛鈧柡鈧?
		client, err := createEsClient(instanceDetail)
		if err != nil {
            logs.Warning(map[string]interface{}{"instance_id": instanceDetail.ResourceID, "error": err.Error()}, "Elasticsearch instance unavailable, skipping client initialization")
			UnlockEsClients() // 闂佸憡鍨跺浠嬪极婵犲洤绫嶉柡鍫㈡暩閻﹀秹姊婚崶锝呬壕闁荤喐娲戝ù鍥暰闂?
			continue
		}

		EsClients[instanceDetail.ResourceID] = client // 闁诲繐绻愬Λ妤咁敇瑜版帒绠ｅ瀣瘨娴煎倿鎮楀☉娅亪宕戝澶婄?esClients
        logs.Info(map[string]interface{}{"instance_id": instanceDetail.ResourceID}, "Elasticsearch client created")

		UnlockEsClients() // 闁荤喐鐟辩紞渚€寮?
	}
}

// 闂佸憡甯楃粙鎴犵磽?es 闁诲骸绠嶉崹娲春濞戞氨鍗?
func createEsClient(instanceDetail dal.ResourceDetail) (*elasticsearch.Client, error) {
	// 婵炲濮撮惄鐥檇ress闁诲孩绋掗〃鍡涱敊瀹€鍕殧鐎瑰嫭婢樼徊鍧楁煕閿旇姤绶叉繛鐓庡暣閺佸秴鐣濋崟顑跨帛闁荤喐娲戠粈渚€宕㈠☉娆愬妞ゆ帊绀佹惔濠傗槈閹捐櫕鎯堥柣鈯欏懐绠旈柨鏃囧劵椤?
	var addr string
	if instanceDetail.Address != nil {
		addr = *instanceDetail.Address
	} else {
        return nil, fmt.Errorf("instance address is empty")
	}

	if instanceDetail.HttpsEnabled != nil && *instanceDetail.HttpsEnabled == true {
		addr = "https://" + addr
	} else {
		addr = "http://" + addr
	}

	// 闂備焦婢樼粔鍫曟偪閸℃瑢鍋撻獮鍨仾闁糕晜绋撶划鈺咁敍濠婂嫮鏆犻梺鍛婄懃閸婂綊寮?
	cfg := elasticsearch.Config{
		Addresses: []string{addr}, // 闂傚倸妫楀Λ娑㈠礂閵忋倖鍎嶉柛鏇ㄥ亝閸曢箖鏌涜閸?
	}

	// 闁荤喐鐟辩徊楣冩倵閻ｅ本濯奸柕鍫濈墢濡插牓姊洪弶璺ㄐら柣?
	authConfigs := parseAuthConfigs(string(instanceDetail.AuthConfigs))

	for authType, configValue := range authConfigs {
		switch authType {
		case "basic":
			if configValue.ConfigValue != "" {
				raw := strings.TrimSpace(configValue.ConfigValue)
				if strings.HasPrefix(raw, "{") {
					var basicMap map[string]string
					if err := json.Unmarshal([]byte(raw), &basicMap); err == nil {
						cfg.Username = basicMap["username"]
						cfg.Password = basicMap["password"]
						break
					}
				}

				cfg.Username = configValue.ConfigKey
				cfg.Password = configValue.ConfigValue
			} else {
				cfg.Username = configValue.ConfigKey
			}

		case "api_key":
			if configValue.ConfigValue != "" {
				cfg.APIKey = configValue.ConfigValue // API 闁诲酣娼уΛ婵嬪箰?
			}
		}
	}

	// 闁荤姴鎼悿鍥╂崲閸愵亝瀚氬ù锝呭槻婵稑螖閻樿尙鐒烽柣锕€顦甸悰顕€骞嗛柇锕€鏅紓鍌氬枤閸犳稓绮旈幘顔肩睄?
	skipSSL := false
	if instanceDetail.SkipSslVerify != nil {
		skipSSL = *instanceDetail.SkipSslVerify
	}
	cfg.Transport = &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: skipSSL},
		ResponseHeaderTimeout: 10 * time.Second,
	}

	// 闂佸憡甯楃粙鎴犵磽?Elasticsearch 闁诲骸绠嶉崹娲春濞戞氨鍗?
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %v", err)
	}

	// 濠电偞娼欓鍫ユ儊椤栨稒浜ら柣鎰綑婢?
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %v", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.IsError() {
		return nil, fmt.Errorf("failed to get response from Elasticsearch: %v", res.Status())
	}

	return client, nil
}

// 闁荤喐鐟辩徊楣冩倵閻ｅ本濯奸柕鍫濈墢濡插牓姊洪弶璺ㄐら柣?
func parseAuthConfigs(authConfigsStr string) map[string]dal.AuthConfig {
	authConfigs := make(map[string]dal.AuthConfig)

	// 闁荤姳闄嶉崐娑㈡儊婢舵劖鐓€鐎广儱娲ㄩ弸鍌炴煛閸曢潧鐏熷ù婊勫⒋SON闂佸搫绉堕崢褏妲愰敍鍕ㄥ亾濞戞顏堝磻瀹ュ鍎?
	if authConfigsStr == "" {
		return authConfigs
	}

	// 闁诲繐绻戠换鍡涙儊椤栨粍鍠嗛柨婵嗘閳ь剝濮ょ粙澶屽姬閹跨澊N闂佸搫绉堕崢褏妲?
	var jsonConfigs map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(authConfigsStr), &jsonConfigs); err == nil {
		// JSON闂佸搫绉堕崢褏妲愰敍鍕枂闁挎繂妫涢埀顑跨矙楠炲骞囬鈧～?
		for _, configData := range jsonConfigs {
			if _, ok := configData["config_value"].(string); ok {
				realAuthType, _ := configData["auth_type"].(string)
				// 2. 闂佸吋鍎抽崲鑼躲亹閸モ斁鍋撻悽鍨殌缂併劍鐓￠幆鍐礋椤栨艾璧嬬紓鍌氬枤閸犳牠鍩€?
				configValue, _ := configData["config_value"].(string)
				authConfigs[realAuthType] = dal.AuthConfig{
					ConfigKey:   realAuthType,
					ConfigValue: configValue,
				}
			}
		}
		return authConfigs
	}

	// 婵犵鈧啿鈧綊鎮樻繅濉朞N闁荤喐鐟辩徊楣冩倵閼恒儱绶為弶鍫亯琚濋梺鎸庣☉閼活垶鎯冮崜褎瀚氶柡鍥╁枎閻﹀姊洪锝庢當鐟滄澘娼″畷姘跺幢濞戞瑯鍚橀梺姹囧妼鐎氫即寮銏犵９闁绘挸瀵掗崵鐘绘煛瀹ュ洤甯剁紒鎲嬬磿閹叉挳鏁愰崱娆屽亾?
	configPairs := splitTopLevelCommas(authConfigsStr)
	for _, pair := range configPairs {
		colonIndex := strings.Index(pair, ":")
		if colonIndex <= 0 {
			continue
		}
		key := pair[:colonIndex]
		value := pair[colonIndex+1:]

		if strings.HasPrefix(strings.TrimSpace(value), "{") {
			authConfigs[key] = dal.AuthConfig{
				ConfigKey:   key,
				ConfigValue: value,
			}
			continue
		}

		// 闂佸搫鎷嬮崳锝夊焵?key:value 闂佸搫绉堕崢褏妲?
		authConfigs[key] = dal.AuthConfig{
			ConfigKey:   key,
			ConfigValue: value,
		}
	}

	return authConfigs
}

func splitTopLevelCommas(s string) []string {
	var parts []string
	last := 0
	depth := 0
	inQuotes := false
	var prevRune rune
	for i, r := range s {
		if r == '"' && prevRune != '\\' {
			inQuotes = !inQuotes
		}
		if !inQuotes {
			if r == '{' || r == '[' || r == '(' {
				depth++
			} else if r == '}' || r == ']' || r == ')' {
				if depth > 0 {
					depth--
				}
			} else if r == ',' && depth == 0 {
				parts = append(parts, strings.TrimSpace(s[last:i]))
				last = i + 1
			}
		}
		prevRune = r
	}
	if last <= len(s)-1 {
		parts = append(parts, strings.TrimSpace(s[last:]))
	}
	return parts
}

// 闂佺绻戞繛濠偽?es 闁诲骸绠嶉崹娲春濞戞氨鍗氭い鏍ㄧ箘缁犻箖鏌?
func closeEsClient(client *elasticsearch.Client) {
	if client.Transport != nil {
		// 闂佺绻戞繛濠偽涚€涙ɑ浜ら柣鎰綑婢?
		if esTransport, ok := client.Transport.(*elastictransport.Client); ok {
			// 婵炶揪缍€濞夋洟寮妶澶婄煑鐎广儱鎳愬▓?
			rv := reflect.ValueOf(esTransport).Elem()
			transportField := rv.FieldByName("transport")
			if transportField.IsValid() && !transportField.IsZero() {
				// 闂佸吋鍎抽崲鑼躲亹閸ｅ兗tp.Transport
				transportValue := transportField.Interface()
				if httpTransport, ok := transportValue.(*http.Transport); ok {
					//闂佺绻戞繛濠偽涚€靛摜鐭氭繛宸簼閿涙繈寮堕埡鍐ㄤ沪閻?
					httpTransport.CloseIdleConnections()
                    log.Println("Elasticsearch client connection closed")
				} else {
                    log.Println("unable to access underlying http.Transport")
				}
			} else {
                log.Println("unable to access transport field")
			}
		} else {
            log.Println("unexpected Elasticsearch transport type")
		}
	}
}

// CreateEsClient 闁诲繐绻嬪ù鍥夋繝鍥х闁告稒娼欓崝?
func CreateEsClient(authConfig dal.AuthConfig) (*elasticsearch.Client, error) {

	if authConfig.ResourceID == 0 {
		return nil, fmt.Errorf("invalid resource id: %d", authConfig.ResourceID)
	}

	instanceDetail, err := getInstanceDetailByID(authConfig.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query instance detail: %v", err)
	}

	// 婵炶揪缍€濞夋洟寮妶澶婄闁告稒娼欓崝?createEsClient 闂佸憡甯楃粙鎴犵磽閹惧灈鍋撻獮鍨仾闁糕晜绋撶划?
	client, err := createEsClient(*instanceDetail)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getEsClient(instanceID uint) (*elasticsearch.Client, bool) {
	esClientsMutex.RLock()
	defer esClientsMutex.RUnlock()
	c, ok := EsClients[instanceID]
	return c, ok
}

// GetEsClient 闁诲繐绻嬪ù鍥夋繝鍥х闁告稒娼欓崝銉╂煥濞戞ɑ婀伴柛褍锕畷婵嬫偐閸愯尙妯嗛梻鍌氭礌閸嬫捇鏌涢幒鎾剁畵妞ゎ偅鍔欏畷鐘诲冀椤撴稑浜鹃柡鍕箳鐢?
func GetEsClient(instanceID uint) (*elasticsearch.Client, bool) {
	// 闂佺绻愰悧鍡涙儍閸撗勫珰闁哄洦姘ㄩ惌銈囩磽閸屾稒灏柣掳鍔嶇粙澶愵敇閳ュ磭顔嗛梺?
	client, exists := getEsClient(instanceID)
	if exists {
		return client, true
	}

	// 婵犵鈧啿鈧綊鎮樻径瀣枖鐎广儱鎳愰幗鐘绘煕閿旇崵鍘滅紒杈ㄧ箞瀹曟艾鈻庨幇顒傛▎闂傚倸娲犻崑鎾绘煕閹烘挾绠撴い顐ｅ姍瀹?
    logs.Info(map[string]interface{}{"instance_id": instanceID}, "Elasticsearch client missing, initializing on demand")

	// 闂佸吋鍎抽崲鑼躲亹閸モ斁鍋撻崷顓炰粧缂佽翰鍎抽幏鐘绘晜閽樺澹?
	instanceDetail, err := getInstanceDetailByID(instanceID)
	if err != nil {
        logs.Warning(map[string]interface{}{"instance_id": instanceID, "error": err.Error()}, "Failed to load Elasticsearch instance detail")
		return nil, false
	}

	// 闂佸憡甯楃粙鎴犵磽閹惧灈鍋撻獮鍨仾闁糕晜绋撶划?
	client, err = createEsClient(*instanceDetail)
	if err != nil {
        logs.Warning(map[string]interface{}{"instance_id": instanceID, "error": err.Error()}, "Elasticsearch instance unavailable during lazy init")
		return nil, false
	}

	// 闁诲孩绋掗敋闁稿绉瑰畷姘跺箥椤旀儳顦╅柣?
	LockEsClients()
	EsClients[instanceID] = client
	UnlockEsClients()

    logs.Info(map[string]interface{}{"instance_id": instanceID}, "Elasticsearch client initialized on demand")
	return client, true
}

// CloseEsClient 闁诲繐绻嬪ù鍥夋繝鍥х闁告稒娼欓崝?
func CloseEsClient(client *elasticsearch.Client) {
	closeEsClient(client)
}

// LockEsClients 闂佸憡姊绘繛鈧柡鈧?
func LockEsClients() {
	esClientsMutex.Lock()
}

// UnlockEsClients 闁荤喐鐟辩紞渚€寮?
func UnlockEsClients() {
	esClientsMutex.Unlock()
}

// SafeSetEsClient 闂侀潻璐熼崝灞炬叏閻愮儤鐓ュù锝呮惈缁犳艾顭胯閸熲晝绮崨顖涘闁告挆鍛€柣搴＄畭閸ㄦ椽宕哄☉姘卞崥妞ゆ牓鍊楃粈鍕⒒閸ワ絽浜炬繝銏ｅ煐閻楃娀宕曢幘顔肩闁割偓绲垮▓鍫曟煟?LockEsClients闂?
func SafeSetEsClient(instanceID uint, client *elasticsearch.Client) {
	EsClients[instanceID] = client
    logs.Info(map[string]interface{}{"instance_id": instanceID}, "Elasticsearch client set in cache")
}

// SafeDeleteEsClient 闂侀潻璐熼崝灞炬叏閻愮儤鐓ュù锝呮惈缁犳艾顭胯閸熲晝绮崨鏉戠闁绘绮悵鐔兼倵楠炲灝鐏柛鈺傜〒缁晠顢涢妶鍥╊槱闂傚倸娲犻崑鎾愁熆閼哥數澧甸柛搴㈡尦瀹曟宕奸敐鍥啀闂?LockEsClients闂?
func SafeDeleteEsClient(instanceID uint) {
	delete(EsClients, instanceID)
}

// getElasticsearchInstances 闂佸吋鍎抽崲鑼躲亹閸ヮ剙绠ラ柍褜鍓熷鍨箙閸嶇浛sticsearch闁诲骸婀遍崑妯兼?
func getElasticsearchInstances() ([]dal.ResourceDetail, error) {
	// 婵炶揪缍€濞夋洟寮妶澶婂偍闁绘梻鍎ら弲绐糛L闂佸搫琚崕鎾敋濡ゅ懏鐒奸柛顭戝枛鐢磭鈧敻鍋婃禍鐐虹嵁閸℃鐟规繝闈涳功椤?
	var instanceDetails []dal.ResourceDetail
	err := GORMDB.Where("resource_type = ? AND resource_subtype = ?", "instance", "elasticsearch").Find(&instanceDetails).Error
	return instanceDetails, err
}

// getInstanceDetailByID 闂佸搫绉烽～澶婄暤娑擃搳闂佸吋鍎抽崲鑼躲亹閸モ斁鍋撻崷顓炰粧缂佽翰鍎抽幏鐘绘晜閽樺澹?
func getInstanceDetailByID(id uint) (*dal.ResourceDetail, error) {
	var instanceDetail dal.ResourceDetail
	err := GORMDB.Where("resource_id = ? AND resource_type = ?", id, "instance").First(&instanceDetail).Error
	if err != nil {
		return nil, err
	}
	return &instanceDetail, nil
}
