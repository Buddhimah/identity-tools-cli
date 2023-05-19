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
	"text/tabwriter"

	"github.com/wso2-extensions/identity-tools-cli/iamctl/pkg/utils"
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

	if utils.TOOL_CONFIGS.ApplicationConfigs != nil {
		return utils.ResolveAdvancedKeywordMapping(appName, utils.TOOL_CONFIGS.ApplicationConfigs)
	}
	return utils.TOOL_CONFIGS.KeywordMappings
}
