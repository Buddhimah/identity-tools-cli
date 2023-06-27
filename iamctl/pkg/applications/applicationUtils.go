/**
* Copyright (c) 2023, WSO2 LLC. (https://www.wso2.com) All Rights Reserved.
*
* WSO2 LLC. licenses this file to you under the Apache License,
* Version 2.0 (the "License"); you may not use this file except
* in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied. See the License for the
* specific language governing permissions and limitations
* under the License.
 */

package applications

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
	"gopkg.in/yaml.v2"
)

type Application struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AppList struct {
	Applications []Application `json:"applications"`
}

type AppConfig struct {
	ApplicationName string `yaml:"applicationName"`
}

type AuthConfig struct {
	InboundAuthenticationConfig struct {
		InboundAuthenticationRequestConfigs []struct {
			InboundAuthType string `yaml:"inboundAuthType"`
			InboundAuthKey  string `yaml:"inboundAuthKey"`
		} `yaml:"inboundAuthenticationRequestConfigs"`
	} `yaml:"inboundAuthenticationConfig"`
}

func getDeployedAppNames() []string {

	apps := getAppList()
	var appNames []string
	for _, app := range apps {
		appNames = append(appNames, app.Name)
	}
	return appNames
}

func getAppList() (spIdList []Application) {

	var list AppList
	resp, err := utils.SendGetListRequest(utils.APPLICATIONS)
	if err != nil {
		log.Println("Error while retrieving application list", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		writer := new(tabwriter.Writer)
		writer.Init(os.Stdout, 8, 8, 0, '\t', 0)
		defer writer.Flush()

		err = json.Unmarshal(body, &list)
		if err != nil {
			log.Fatalln(err)
		}
		resp.Body.Close()

		spIdList = list.Applications
	} else if error, ok := utils.ErrorCodes[statusCode]; ok {
		log.Println(error)
	} else {
		log.Println("Error while retrieving application list")
	}
	return spIdList
}

func getAppKeywordMapping(appName string) map[string]interface{} {

	if utils.KEYWORD_CONFIGS.ApplicationConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(appName, utils.KEYWORD_CONFIGS.ApplicationConfigs)
	}
	return utils.KEYWORD_CONFIGS.KeywordMappings
}

func isAuthenticationApp(fileData string) (bool, error) {

	config, err := unmarshalAuthConfig([]byte(fileData))
	if err != nil {
		return false, err
	}

	for _, requestConfig := range config.InboundAuthenticationConfig.InboundAuthenticationRequestConfigs {
		if strings.ToLower(requestConfig.InboundAuthType) == utils.OAUTH2 {
			return true, nil
		}
	}
	return false, nil
}

func checkInboundAuthKey(fileData []byte) (bool, error) {

	config, err := unmarshalAuthConfig(fileData)
	if err != nil {
		return false, err
	}

	for _, requestConfig := range config.InboundAuthenticationConfig.InboundAuthenticationRequestConfigs {
		if requestConfig.InboundAuthKey == utils.SERVER_CONFIGS.ClientId {
			return true, nil
		}
	}
	return false, nil
}

func unmarshalAuthConfig(data []byte) (AuthConfig, error) {
	var config AuthConfig
	err := yaml.Unmarshal(data, &config)
	return config, err
}

func maskOAuthConsumerSecret(fileContent []byte) []byte {

	// Find and replace the value of oauthConsumerSecret with asterisks
	maskedValue := "'********'"
	pattern := "(?m)(^\\s*oauthConsumerSecret:\\s*)null\\s*$"
	re := regexp.MustCompile(pattern)
	maskedContent := re.ReplaceAllString(string(fileContent), "${1}"+maskedValue)

	return []byte(maskedContent)
}

func IsManagementApp(file os.FileInfo, importFilePath string) bool {

	appFilePath := filepath.Join(importFilePath, file.Name())
	fileData, err := ioutil.ReadFile(appFilePath)
	if err != nil {
		log.Printf("Error reading file: %s\n", err.Error())
		return false
	}
	isManagement, err := checkInboundAuthKey(fileData)
	if err != nil {
		log.Printf("Error checking if file is a management app: %s\n", err.Error())
		return false
	}
	if isManagement {
		appName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		log.Printf("Info: Management App: %s is excluded from deletion.\n", appName)
		return true
	}
	return false
}
