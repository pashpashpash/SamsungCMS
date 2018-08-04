package main
import (
    "log"
    "strconv"
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
featuredLocations := []string{}
configurationOrder := 0


//For [ALL COUNTRIES] section in export
homescreenConfigsInAllCountries := []string{}
folderConfigsInAllCountries := []string{}
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

generateConfigQuery = `SELECT DISTINCT featuredLocationName from featuredLocations`
log.Println("generateConfigurationINI –\t\tQuery = " + generateConfigQuery)
featuredLocationList, err := db.Query(generateConfigQuery)
checkErr(err)
for(featuredLocationList.Next()){
    featuredLocationName := string("")
    featuredLocationList.Scan(&featuredLocationName)
    featuredLocations = append(featuredLocations, featuredLocationName)
}

for _, config := range configurationNumbers {
    for _, featuredLocation := range featuredLocations {
        generateConfigQuery = `
        SELECT * FROM (SELECT DISTINCT Country_ID from countries)
        EXCEPT
        SELECT * FROM (SELECT DISTINCT Country_ID from configurationMappings
        INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
        INNER JOIN featuredLocations ON appConfigs.Config_ID = featuredLocations.Config_ID
        WHERE featuredLocations.featuredLocationName = "`+featuredLocation+`"
        AND configurationMappings.Config_ID = "`+config+`")
        `
        countriesList, err := db.Query(generateConfigQuery)
        checkErr(err)
        existsEverywhere := true;
        for(countriesList.Next()) {
            existsEverywhere = false;
        }
        if(existsEverywhere) {
            if(featuredLocation == "homescreen") {
                homescreenConfigsInAllCountries = append(homescreenConfigsInAllCountries, config)
            } else if (featuredLocation == "folder") {
                folderConfigsInAllCountries = append(folderConfigsInAllCountries, config)
            } else if (featuredLocation == "max") {
                maxConfigsInAllCountries = append(maxConfigsInAllCountries, config)
            } else if (featuredLocation == "maxGo") {
                maxGoConfigsInAllCountries = append(maxGoConfigsInAllCountries, config)
            }
        }
    }
}


homescreenConfigs := difference(configurationNumbers, homescreenConfigsInAllCountries) //used in the countries section
folderConfigs := difference(configurationNumbers, folderConfigsInAllCountries)
maxConfigs := difference(configurationNumbers, maxConfigsInAllCountries)
maxGoConfigs := difference(configurationNumbers, maxGoConfigsInAllCountries)


log.Println("Countries:")
log.Println(countries)
log.Println("Operators:")
log.Println(operatorGroups)
log.Println("Configuration Numbers:")
log.Println(configurationNumbers)
log.Println("Featured Locations:")
log.Println(featuredLocations)
log.Println("Configs In all Countries (homescreen,folder,max,maxGo):")
log.Println(homescreenConfigsInAllCountries)
log.Println(folderConfigsInAllCountries)
log.Println(maxConfigsInAllCountries)
log.Println(maxGoConfigsInAllCountries)
log.Println("Configs to be iterated over in [countries] and [operatorGroups] sections in config (homescreen,folder,max,maxGo):")
log.Println(homescreenConfigs)
log.Println(folderConfigs)
log.Println(maxConfigs)
log.Println(maxGoConfigs)

configuration += "\n; ========================  Defaults  ======================================"
for _, globalfeaturedLocation := range featuredLocations { //featuredLocation = product i.e. "maxGo" or "folder"
    configuration += (("\n["+"ALLCOUNTRIES"+"_"+globalfeaturedLocation+"]") + ("\n"))
    configuration += (("order = " +  strconv.Itoa(configurationOrder)) + ("\n"))
    configuration += "filter = (["
        configuration += "\"product\": \""+globalfeaturedLocation+"\","
    configuration += ("])\n")
    configuration += ("configList = [")
    if(globalfeaturedLocation == "homescreen") {
        for _, globalAppConfig := range  homescreenConfigsInAllCountries{
            configuration += (globalAppConfig + ", ")
        }
    } else if(globalfeaturedLocation == "folder") {
        for _, globalAppConfig := range  folderConfigsInAllCountries{
            configuration += (globalAppConfig + ", ")
        }
    } else if(globalfeaturedLocation == "max") {
        for _, globalAppConfig := range  maxConfigsInAllCountries{
            configuration += (globalAppConfig + ", ")
        }
    }  else if(globalfeaturedLocation == "maxGo") {
        for _, globalAppConfig := range maxGoConfigsInAllCountries {
            configuration += (globalAppConfig + ", ")
        }
    }
    configuration += ("]")
    configurationOrder++
}


configuration += "\n; ========================  Countries  ======================================"
log.Println("Starting to dump countries...")
for _, country := range countries {
    // log.Println(country)



    for _, featuredLocation := range featuredLocations { //featuredLocation = product i.e. "maxGo" or "folder"
        configuration += (("\n["+country+"_"+featuredLocation+"]") + ("\n"))
        configuration += (("order = " +  strconv.Itoa(configurationOrder)) + ("\n"))
        configuration += "filter = (["
            configuration += "\"country\": \""+country+"\","
            configuration += "\"product\": \""+featuredLocation+"\","
        configuration += ("])\n")
        configuration += ("configList = [")
        configs := []string{}
        var configsMap = make(map[string]bool)
        if(featuredLocation == "homescreen") {
            for _, appConfig := range homescreenConfigs {
                //check if app configuration exists based on config number, featuredLocation, and country_id
                orderOfConfig := string("")
                orderOfConfig = configExistsInCountryAndProduct(country, featuredLocation, appConfig)
                if(orderOfConfig != "") {
                    if(!(thereAreSimilarCountryConfigsWithHigherOrder(orderOfConfig, country, featuredLocation, appConfig))) {
                        //only add config if this is the highest order instance of the config in this section with this originalName
                            addUnique(&configs, configsMap, appConfig)

                    }
                }
            }
            configs = RemoveDuplicatesFromSlice(configs)

        } else if (featuredLocation == "folder") {
            for _, appConfig := range folderConfigs {
                //check if app configuration exists based on config number, featuredLocation, and country_id
                //change this to get countryOrder as well
                orderOfConfig := string("")
                orderOfConfig = configExistsInCountryAndProduct(country, featuredLocation, appConfig)
                if(orderOfConfig != "") {

                    //add app config to country
                    if(!(thereAreSimilarCountryConfigsWithHigherOrder(orderOfConfig, country, featuredLocation, appConfig))) {
                        //only add config if this is the highest order instance of the config in this section with this originalName
                        addUnique(&configs, configsMap, appConfig)
                    }
                }
            }
            configs = RemoveDuplicatesFromSlice(configs)

        } else if (featuredLocation == "max") {
            for _, appConfig := range maxConfigs {
                //check if app configuration exists based on config number, featuredLocation, and country_id
                orderOfConfig := string("")
                orderOfConfig = configExistsInCountryAndProduct(country, featuredLocation, appConfig)
                if(orderOfConfig!= "") {

                    if(!(thereAreSimilarCountryConfigsWithHigherOrder(orderOfConfig, country, featuredLocation, appConfig))) {
                        //only add config if this is the highest order instance of the config in this section with this originalName
                        addUnique(&configs, configsMap, appConfig)
                    }
                }
            }
            configs = RemoveDuplicatesFromSlice(configs)

        } else if (featuredLocation == "maxGo") {
            for _, appConfig := range maxGoConfigs {
                //check if app configuration exists based on config number, featuredLocation, and country_id
                orderOfConfig := string("")
                orderOfConfig = configExistsInCountryAndProduct(country, featuredLocation, appConfig)

                if(orderOfConfig != "") {

                    if(!(thereAreSimilarCountryConfigsWithHigherOrder(orderOfConfig, country, featuredLocation, appConfig))) {
                        //only add config if this is the highest order instance of the config in this section with this originalName
                        addUnique(&configs, configsMap, appConfig)
                    }
                }
            }
            configs = RemoveDuplicatesFromSlice(configs)

        }
        for _, config := range configs {
            configuration += (config + ", ")
        }
        configuration += ("]")
        configurationOrder++
    }

}
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
    for _, featuredLocation := range featuredLocations { //featuredLocation = product i.e. "maxGo" or "folder"
        configuration += (("\n["+operatorGroup+"_"+featuredLocation+"]") + ("\n"))
        configuration += (("order = " +  strconv.Itoa(configurationOrder)) + ("\n"))
        configuration += "filter = (["
            configuration += "\"operator\": \""+operatorGroup+"\","
            configuration += "\"product\": \""+featuredLocation+"\","
        configuration += ("])\n")
        configuration += ("configList = [")
        configs := []string{}
        var configsMap = make(map[string]bool)
        for _, mappedOperator := range mappedOperators {
            // log.Println(mappedOperator)
            if(featuredLocation == "homescreen") {
                for _, appConfig := range homescreenConfigs {
                    results := configExistsInOperatorAndProduct(mappedOperator, featuredLocation, appConfig)
                    operatorOrder := results[0]
                    country_id := results[1]

                    countryOrder := getCountryMappingOrder(country_id, featuredLocation, appConfig)
                    countryOrderInt, _ := strconv.Atoi(countryOrder)
                    operatorOrderInt, _ := strconv.Atoi(operatorOrder)
                    if(countryOrderInt < operatorOrderInt) {

                        if(!(thereAreSimilarOperatorConfigsWithHigherOrder(operatorOrder, mappedOperator, featuredLocation, appConfig))) {
                            addUnique(&configs, configsMap, appConfig)
                        }
                    }
                }
                configs = RemoveDuplicatesFromSlice(configs)
            } else if(featuredLocation == "folder") {
                for _, appConfig := range folderConfigs {
                    results := configExistsInOperatorAndProduct(mappedOperator, featuredLocation, appConfig)
                    operatorOrder := results[0]
                    country_id := results[1]

                    countryOrder := getCountryMappingOrder(country_id, featuredLocation, appConfig)
                    countryOrderInt, _ := strconv.Atoi(countryOrder)
                    operatorOrderInt, _ := strconv.Atoi(operatorOrder)
                    if(countryOrderInt < operatorOrderInt) {

                        if(!(thereAreSimilarOperatorConfigsWithHigherOrder(operatorOrder, mappedOperator, featuredLocation, appConfig))) {
                            addUnique(&configs, configsMap, appConfig)
                        }
                    }
                }
                configs = RemoveDuplicatesFromSlice(configs)
            }  else if(featuredLocation == "max") {
                for _, appConfig := range maxConfigs {
                    results := configExistsInOperatorAndProduct(mappedOperator, featuredLocation, appConfig)
                    operatorOrder := results[0]
                    country_id := results[1]

                    countryOrder := getCountryMappingOrder(country_id, featuredLocation, appConfig)
                    countryOrderInt, _ := strconv.Atoi(countryOrder)
                    operatorOrderInt, _ := strconv.Atoi(operatorOrder)
                    if(countryOrderInt < operatorOrderInt) {

                        if(!(thereAreSimilarOperatorConfigsWithHigherOrder(operatorOrder, mappedOperator, featuredLocation, appConfig))) {
                            addUnique(&configs, configsMap, appConfig)
                        }
                    }
                }
                configs = RemoveDuplicatesFromSlice(configs)
            }  else if(featuredLocation == "maxGo") {
                for _, appConfig := range maxGoConfigs {
                    results := configExistsInOperatorAndProduct(mappedOperator, featuredLocation, appConfig)
                    operatorOrder := results[0]
                    country_id := results[1]

                    countryOrder := getCountryMappingOrder(country_id, featuredLocation, appConfig)
                    countryOrderInt, _ := strconv.Atoi(countryOrder)
                    operatorOrderInt, _ := strconv.Atoi(operatorOrder)
                    if(countryOrderInt < operatorOrderInt) {

                        if(!(thereAreSimilarOperatorConfigsWithHigherOrder(operatorOrder, mappedOperator, featuredLocation, appConfig))) {
                            addUnique(&configs, configsMap, appConfig)
                        }
                    }
                }
                configs = RemoveDuplicatesFromSlice(configs)

            }
        }
        configurationOrder++
        for _, config := range configs {
            configuration += (config + ", ")
        }
        configuration += ("]")
    }
}
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

func configExistsInCountryAndProduct(country, featuredLocation, appConfig  string) (string) {
    generateConfigQuery := `
    SELECT DISTINCT configurationMappings.id from configurationMappings
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN featuredLocations ON appConfigs.Config_ID = featuredLocations.Config_ID
    WHERE configurationMappings.Country_ID = '`+country+`' AND featuredLocations.featuredLocationName = "`+featuredLocation+`"
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

func configExistsInOperatorAndProduct(operator, featuredLocation, appConfig  string) ([]string) {
    generateConfigQuery := `
    SELECT DISTINCT operators.Country_ID, configurationMappings.id from configurationMappings
    JOIN operators USING (MCCMNC_ID)
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN featuredLocations ON appConfigs.Config_ID = featuredLocations.Config_ID
    WHERE MCCMNC_ID = '`+operator+`' AND configurationMappings.Config_ID = "`+appConfig+`" AND featuredLocations.featuredLocationName ="`+featuredLocation+`" ORDER BY configurationMappings.id DESC LIMIT 1
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
func thereAreSimilarCountryConfigsWithHigherOrder(order, country, featuredLocation, appConfig  string) (bool) {
    generateConfigQuery := `
    SELECT DISTINCT configurationMappings.id from configurationMappings
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN featuredLocations ON appConfigs.Config_ID = featuredLocations.Config_ID
    WHERE configurationMappings.Country_ID = '`+country+`' AND featuredLocations.featuredLocationName = "`+featuredLocation+`"
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
func thereAreSimilarOperatorConfigsWithHigherOrder(order, operator, featuredLocation, appConfig  string) (bool) {
    generateConfigQuery := `
    SELECT DISTINCT configurationMappings.id from configurationMappings
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN featuredLocations ON appConfigs.Config_ID = featuredLocations.Config_ID
    WHERE configurationMappings.MCCMNC_ID = '`+operator+`' AND featuredLocations.featuredLocationName = "`+featuredLocation+`"
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

func getCountryMappingOrder(country, featuredLocation, appConfig  string) (string) {
    generateConfigQuery := `
    SELECT DISTINCT configurationMappings.id, configurationMappings.Config_ID from configurationMappings
    INNER JOIN appConfigs ON configurationMappings.Config_ID = appConfigs.Config_ID
    INNER JOIN featuredLocations ON appConfigs.Config_ID = featuredLocations.Config_ID
    WHERE configurationMappings.Country_ID = '`+country+`' AND featuredLocations.featuredLocationName = "`+featuredLocation+`"
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
                      log.Println(item + " is a duplicate")
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
