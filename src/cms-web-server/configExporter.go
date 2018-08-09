package main
import (
    "log"
    "strconv"
    "archive/zip"
    "path/filepath"
    "strings"
    "os"
    "io"
)
func generateConfigurationINI() (string) {

configuration := string("")

configuration +=
`
[global]
known filters = ({
    "country",
    "os_version",
    "device_model",
    "operator"
})
`
//Will use these for export
countries := []string{}
operatorGroups := []string{}
configurationNumbers := []string{}
products := []string{}
configurationOrder := 0


//For [ALL COUNTRIES] section in export
maxGlobalConfigsInAllCountries := []string{}
maxConfigsInAllCountries := []string{}
maxGoConfigsInAllCountries := []string{}





generateConfigQuery := `SELECT DISTINCT Country_ID FROM countries ORDER BY name`
log.Println("generateConfigurationINI –\t\tQuery = " + generateConfigQuery)
countryList, err := db.Query(generateConfigQuery)
checkErr(err)
for(countryList.Next()){
    countryRow := string("")
    countryList.Scan(&countryRow)
    countries = append(countries, countryRow)
}

generateConfigQuery = `SELECT DISTINCT Operator_Group_Name from operatorGroups`
log.Println("generateConfigurationINI –\t\tQuery = " + generateConfigQuery)
operatorGroupList, err := db.Query(generateConfigQuery)
checkErr(err)
for(operatorGroupList.Next()){
    group := string("")
    operatorGroupList.Scan(&group)
    operatorGroups = append(operatorGroups, group)
}

generateConfigQuery = `SELECT DISTINCT Config_ID from appConfigs`
log.Println("generateConfigurationINI –\t\tQuery = " + generateConfigQuery)
configList, err := db.Query(generateConfigQuery)
checkErr(err)
for(configList.Next()){
    config := string("")
    configList.Scan(&config)
    configurationNumbers = append(configurationNumbers, config)
}

generateConfigQuery = `SELECT DISTINCT productName from products`
log.Println("generateConfigurationINI –\t\tQuery = " + generateConfigQuery)
productList, err := db.Query(generateConfigQuery)
checkErr(err)
for(productList.Next()){
    productName := string("")
    productList.Scan(&productName)
    products = append(products, productName)
}

for _, config := range configurationNumbers {
    for _, product := range products {
        generateConfigQuery = `
        SELECT * FROM (SELECT DISTINCT Country_ID from countries)
        EXCEPT
        SELECT * FROM (SELECT DISTINCT Country_ID from configurationMappings
        INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
        INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
        WHERE products.productName = "`+product+`"
        AND configurationMappings.Config_ID = "`+config+`")
        `
        countriesList, err := db.Query(generateConfigQuery)
        checkErr(err)
        existsEverywhere := true;
        for(countriesList.Next()) {
            existsEverywhere = false;
        }
        if(existsEverywhere) {
            if(product == "maxGlobal") {
                maxGlobalConfigsInAllCountries = append(maxGlobalConfigsInAllCountries, config)
            } else if (product == "max") {
                maxConfigsInAllCountries = append(maxConfigsInAllCountries, config)
            } else if (product == "maxGo") {
                maxGoConfigsInAllCountries = append(maxGoConfigsInAllCountries, config)
            }
        }
    }
}


maxGlobalConfigs := difference(configurationNumbers, maxGlobalConfigsInAllCountries) //used in the countries section
maxConfigs := difference(configurationNumbers, maxConfigsInAllCountries)
maxGoConfigs := difference(configurationNumbers, maxGoConfigsInAllCountries)


log.Println("Countries:")
log.Println(countries)
log.Println("Operators:")
log.Println(operatorGroups)
log.Println("Configuration Numbers:")
log.Println(configurationNumbers)
log.Println("Featured Locations:")
log.Println(products)
log.Println("Configs In all Countries (maxGlobal,max,maxGo):")
log.Println(maxGlobalConfigsInAllCountries)
log.Println(maxConfigsInAllCountries)
log.Println(maxGoConfigsInAllCountries)
log.Println("Configs to be iterated over in [countries] and [operatorGroups] sections in config (maxGlobal,max,maxGo):")
log.Println(maxGlobalConfigs)
log.Println(maxConfigs)
log.Println(maxGoConfigs)

configuration += "\n; ========================  Defaults  ======================================"
for _, globalproduct := range products { //product = product i.e. "maxGo" or "maxGlobal"
    configuration += (("\n["+"ALLCOUNTRIES"+"_"+globalproduct+"]") + ("\n"))
    configuration += (("order = " +  strconv.Itoa(configurationOrder)) + ("\n"))
    configuration += "filter = (["
        configuration += "\"product\": \""+globalproduct+"\","
    configuration += ("])\n")
    configuration += ("configList = [")
    if(globalproduct == "maxGlobal") {
        for _, globalAppConfig := range  maxGlobalConfigsInAllCountries{
            configuration += (globalAppConfig + ", ")
        }
    } else if(globalproduct == "max") {
        for _, globalAppConfig := range  maxConfigsInAllCountries{
            configuration += (globalAppConfig + ", ")
        }
    }  else if(globalproduct == "maxGo") {
        for _, globalAppConfig := range maxGoConfigsInAllCountries {
            configuration += (globalAppConfig + ", ")
        }
    }
    configuration += ("]")
    configurationOrder++
}


configuration += "\n; ========================  Countries  ======================================"
countriesSection := string("")
log.Println("Starting to dump countries...")
for _, country := range countries {
    // log.Println(country)


    for _, product := range products { //product = product i.e. "maxGo" or "maxGlobal"
        configSection := string("")
        configSection += (("\n["+country+"_"+product+"]") + ("\n"))
        configSection += (("order = " +  strconv.Itoa(configurationOrder)) + ("\n"))
        configSection += "filter = (["
            configSection += "\"country\": \""+country+"\","
            configSection += "\"product\": \""+product+"\","
        configSection += ("])\n")
        configSection += ("configList = [")
        configs := []string{}
        var configsMap = make(map[string]bool)
        if(product == "maxGlobal") {
            for _, appConfig := range maxGlobalConfigs {
                //check if app configuration exists based on config number, product, and country_id
                orderOfConfig := string("")
                orderOfConfig = configExistsInCountryAndProduct(country, product, appConfig)
                if(orderOfConfig != "") {
                    if(!(thereAreSimilarCountryConfigsWithHigherOrder(orderOfConfig, country, product, appConfig))) {
                        //only add config if this is the highest order instance of the config in this section with this originalName
                            addUnique(&configs, configsMap, appConfig)

                    }
                }
            }
            configs = RemoveDuplicatesFromSlice(configs)

        } else if (product == "max") {
            for _, appConfig := range maxConfigs {
                //check if app configuration exists based on config number, product, and country_id
                orderOfConfig := string("")
                orderOfConfig = configExistsInCountryAndProduct(country, product, appConfig)
                if(orderOfConfig!= "") {

                    if(!(thereAreSimilarCountryConfigsWithHigherOrder(orderOfConfig, country, product, appConfig))) {
                        //only add config if this is the highest order instance of the config in this section with this originalName
                        addUnique(&configs, configsMap, appConfig)
                    }
                }
            }
            configs = RemoveDuplicatesFromSlice(configs)

        } else if (product == "maxGo") {
            for _, appConfig := range maxGoConfigs {
                //check if app configuration exists based on config number, product, and country_id
                orderOfConfig := string("")
                orderOfConfig = configExistsInCountryAndProduct(country, product, appConfig)

                if(orderOfConfig != "") {

                    if(!(thereAreSimilarCountryConfigsWithHigherOrder(orderOfConfig, country, product, appConfig))) {
                        //only add config if this is the highest order instance of the config in this section with this originalName
                        addUnique(&configs, configsMap, appConfig)
                    }
                }
            }
            configs = RemoveDuplicatesFromSlice(configs)

        }
        if(len(configs) != 0) {  //if mapped configuration length is not 0, add to export
            for _, config := range configs {
                configSection += (config + ", ")
            }
            configSection += ("]")
            configurationOrder++
            countriesSection += configSection
        }
    }

}
configuration += countriesSection //add countriesSection to configuration export

configuration += "\n; ========================  Operators  ======================================"
log.Println("Starting to dump operators (more checks involved here)...")




operatorGroups = []string{}
generateConfigQuery = `
SELECT DISTINCT Operator_Group_Name from operatorGroups
`
operatorGroupsList, err := db.Query(generateConfigQuery)
checkErr(err)
for(operatorGroupsList.Next()) {
    operatorGroup := string("")
    operatorGroupsList.Scan(&operatorGroup)
    operatorGroups = append(operatorGroups, operatorGroup)
}
operatorGroupsList.Close()

operatorsSection := string("")
for _, operatorGroup := range operatorGroups {


    mappedOperators := []string{}
    generateConfigQuery = `
    SELECT DISTINCT MCCMNC_ID from configurationMappings
    JOIN operatorGroups USING (MCCMNC_ID)
    WHERE Operator_Group_Name = "`+operatorGroup+`"
    `
    mappedOperatorList, err := db.Query(generateConfigQuery)
    checkErr(err)
    for(mappedOperatorList.Next()) {
        mappedOperator := string("")
        mappedOperatorList.Scan(&mappedOperator)
        mappedOperators = append(mappedOperators, mappedOperator)
    }
    mappedOperatorList.Close()
    for _, product := range products { //product = product i.e. "maxGo" or "maxGlobal"
        configSection := string("")
        configSection += (("\n["+operatorGroup+"_"+product+"]") + ("\n"))
        configSection += (("order = " +  strconv.Itoa(configurationOrder)) + ("\n"))
        configSection += "filter = (["
            configSection += "\"operator\": \""+operatorGroup+"\","
            configSection += "\"product\": \""+product+"\","
        configSection += ("])\n")
        configSection += ("configList = [")
        configs := []string{}
        var configsMap = make(map[string]bool)
        for _, mappedOperator := range mappedOperators {
            // log.Println(mappedOperator)
            if(product == "maxGlobal") {
                for _, appConfig := range maxGlobalConfigs {
                    results := configExistsInOperatorAndProduct(mappedOperator, product, appConfig)
                    operatorOrder := results[0]
                    country_id := results[1]

                    countryOrder := getCountryMappingOrder(country_id, product, appConfig)
                    countryOrderInt, _ := strconv.Atoi(countryOrder)
                    operatorOrderInt, _ := strconv.Atoi(operatorOrder)
                    if(countryOrderInt < operatorOrderInt) {

                        if(!(thereAreSimilarOperatorConfigsWithHigherOrder(operatorOrder, mappedOperator, product, appConfig))) {
                            addUnique(&configs, configsMap, appConfig)
                        }
                    }
                }
                configs = RemoveDuplicatesFromSlice(configs)
            } else if(product == "max") {
                for _, appConfig := range maxConfigs {
                    results := configExistsInOperatorAndProduct(mappedOperator, product, appConfig)
                    operatorOrder := results[0]
                    country_id := results[1]

                    countryOrder := getCountryMappingOrder(country_id, product, appConfig)
                    countryOrderInt, _ := strconv.Atoi(countryOrder)
                    operatorOrderInt, _ := strconv.Atoi(operatorOrder)
                    if(countryOrderInt < operatorOrderInt) {

                        if(!(thereAreSimilarOperatorConfigsWithHigherOrder(operatorOrder, mappedOperator, product, appConfig))) {
                            addUnique(&configs, configsMap, appConfig)
                        }
                    }
                }
                configs = RemoveDuplicatesFromSlice(configs)
            }  else if(product == "maxGo") {
                for _, appConfig := range maxGoConfigs {
                    results := configExistsInOperatorAndProduct(mappedOperator, product, appConfig)
                    operatorOrder := results[0]
                    country_id := results[1]

                    countryOrder := getCountryMappingOrder(country_id, product, appConfig)
                    countryOrderInt, _ := strconv.Atoi(countryOrder)
                    operatorOrderInt, _ := strconv.Atoi(operatorOrder)
                    if(countryOrderInt < operatorOrderInt) {

                        if(!(thereAreSimilarOperatorConfigsWithHigherOrder(operatorOrder, mappedOperator, product, appConfig))) {
                            addUnique(&configs, configsMap, appConfig)
                        }
                    }
                }
                configs = RemoveDuplicatesFromSlice(configs)

            }
        }

        if(len(configs) != 0) { //if mapped configuration length is not 0, add to export
            for _, config := range configs {
                configSection += (config + ", ")
            }
            configSection += ("]")
            configurationOrder++
            operatorsSection += configSection
        }
    }
}
configuration += operatorsSection //add countriesSection to configuration export

return configuration
}

// difference returns the elements in a that aren't in b
func difference(a, b []string) []string {
    mb := map[string]bool{}
    for _, x := range b {
        mb[x] = true
    }
    ab := []string{}
    for _, x := range a {
        if _, ok := mb[x]; !ok {
            ab = append(ab, x)
        }
    }
    return ab
}

func configExistsInCountryAndProduct(country, product, appConfig  string) (string) {
    generateConfigQuery := `
    SELECT DISTINCT configurationMappings.id from configurationMappings
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
    WHERE configurationMappings.Country_ID = '`+country+`' AND products.productName = "`+product+`"
    AND configurationMappings.Config_ID = "`+appConfig+`" ORDER BY configurationMappings.id DESC LIMIT 1
    `
    config, err := db.Query(generateConfigQuery)
    checkErr(err)
    string1 := string("")
    for(config.Next()){
        config.Scan(&string1)
    }
    return string1
}

func configExistsInOperatorAndProduct(operator, product, appConfig  string) ([]string) {
    generateConfigQuery := `
    SELECT DISTINCT operators.Country_ID, configurationMappings.id from configurationMappings
    JOIN operators USING (MCCMNC_ID)
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
    WHERE MCCMNC_ID = '`+operator+`' AND configurationMappings.Config_ID = "`+appConfig+`" AND products.productName ="`+product+`" ORDER BY configurationMappings.id DESC LIMIT 1
    `
    configList, err := db.Query(generateConfigQuery)
    checkErr(err)
    order := string("")
    Country_ID := string("")
    for configList.Next() {

        configList.Scan(&Country_ID, &order)
    }
    arrayOfStringsArrays := []string{order, Country_ID}
    return arrayOfStringsArrays
}
func thereAreSimilarCountryConfigsWithHigherOrder(order, country, product, appConfig  string) (bool) {
    generateConfigQuery := `
    SELECT DISTINCT configurationMappings.id from configurationMappings
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
    WHERE configurationMappings.Country_ID = '`+country+`' AND products.productName = "`+product+`"
    AND configurationMappings.Config_ID in (SELECT Config_ID FROM appConfigs WHERE originalName in (SELECT originalName from appConfigs WHERE Config_ID ='`+appConfig+`')) ORDER BY configurationMappings.id DESC LIMIT 1
    `
    configList, err := db.Query(generateConfigQuery)
    checkErr(err)
    higherorder := true
    new_order := string("")
    for configList.Next() {
        configList.Scan(&new_order)
    }
    if(new_order!="") {
        new_order_int, _ :=  strconv.Atoi(new_order)
        old_order_int, _ :=  strconv.Atoi(order)
        if(old_order_int>=new_order_int) {

            higherorder = false
        }
    }

    return higherorder
}
func thereAreSimilarOperatorConfigsWithHigherOrder(order, operator, product, appConfig  string) (bool) {
    generateConfigQuery := `
    SELECT DISTINCT configurationMappings.id from configurationMappings
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
    WHERE configurationMappings.MCCMNC_ID = '`+operator+`' AND products.productName = "`+product+`"
    AND configurationMappings.Config_ID in (SELECT Config_ID FROM appConfigs WHERE originalName in (SELECT originalName from appConfigs WHERE Config_ID ='`+appConfig+`')) ORDER BY configurationMappings.id DESC LIMIT 1
    `
    configList, err := db.Query(generateConfigQuery)
    checkErr(err)
    higherorder := true
    new_order := string("")
    for configList.Next() {
        configList.Scan(&new_order)
    }
    new_order_int, _ :=  strconv.Atoi(new_order)
    old_order_int, _ :=  strconv.Atoi(order)
    if(old_order_int>=new_order_int) {
        higherorder = false
    }

    return higherorder
}

func getCountryMappingOrder(country, product, appConfig  string) (string) {
    generateConfigQuery := `
    SELECT DISTINCT configurationMappings.id, configurationMappings.Config_ID from configurationMappings
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN products ON appConfigs.Config_ID = products.Config_ID
    WHERE configurationMappings.Country_ID = '`+country+`' AND products.productName = "`+product+`"
    AND configurationMappings.Config_ID in  (SELECT Config_ID from appConfigs WHERE originalName in (SELECT originalName from appConfigs WHERE Config_ID = '`+appConfig+`')) ORDER BY configurationMappings.id DESC LIMIT 1
    `
    config, err := db.Query(generateConfigQuery)
    checkErr(err)
    id := string("")
    for(config.Next()){
        config.Scan(&id)
    }
    return id
}

func addUnique(a *[]string, m map[string]bool, s string) {
    if m[s] {
        return // Already in the map
    }
    *a = append(*a, s)
    m[s] = true
}

func RemoveDuplicatesFromSlice(s []string) []string {
      m := make(map[string]bool)
      for _, item := range s {
              if _, ok := m[item]; ok {
                      // duplicate item
              } else {
                      m[item] = true
              }
      }

      var result []string
      for item, _ := range m {
              result = append(result, item)
      }
      return result
}

func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
