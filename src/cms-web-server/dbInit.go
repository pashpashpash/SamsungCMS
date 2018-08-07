package main
import ("strings"
    "database/sql"
    "text/tabwriter"
	"os"
	"log")

var db (*sql.DB)

func uniqueNonEmptyElementsOf(s []string) []string {
  unique := make(map[string]bool, len(s))
	us := make([]string, len(unique))
	for _, elem := range s {
		if len(elem) != 0 {
			if !unique[elem] {
				us = append(us, elem)
				unique[elem] = true
			}
		}
	}

	return us
}

// DATABASE HELPER FUNCTION //
func initDB(name string) (*sql.DB) {
    tw := new(tabwriter.Writer)
    tw.Init(os.Stderr, 0, 8, 0, '\t', 0)
    log.Println("initDB –\t\tInitializing SQLite db with the name " + name)
    db, err := sql.Open("sqlite3", "./"+name+".db")
    checkErr(err)

    //===============================================//
    log.Println( "initDB –\t\tcreating SQLite tables")
    createTables(db)

    statement, _ := db.Prepare(`UPDATE mytable SET MCCMNC_ID = mcc||""||mnc`)
    _, err = statement.Exec()
    checkErr(err)

    log.Println("initDB –\t\tInitializing users table with admin username and password...")
    _, err = db.Exec(`INSERT or IGNORE  INTO users (username, password) VALUES ("admin", "admin")`)
    checkErr(err)


    // should be changed to initialize only OUR list of operators
    log.Println("initDB –\t\tInitializing operators table with temporary MCC table data...")
    _, err = db.Exec(`INSERT or IGNORE  INTO operators (MCCMNC_ID, Operator_Name, Country_ID, Operator_Group_Name) SELECT CAST(mytable.MCCMNC_ID AS TEXT), Operator_Name, Country_ID, mytable2.Operator_Group_Name FROM mytable INNER JOIN mytable2 ON mytable.MCCMNC_ID = mytable2.MCCMNC_ID`)
    checkErr(err)

    log.Println("initDB –\t\tExtracting Group_Name column from operators table and respective MCCMNC_ID and putting into operatorGroups table")
    _, err = db.Exec(`INSERT or IGNORE  INTO operatorGroups (Operator_Group_Name, MCCMNC_ID) SELECT operators.Operator_Group_Name, MCCMNC_ID FROM operators`)
    checkErr(err)

    log.Println("initDB –\t\tDeleting Operator_Group_Name column from operators table")
    _, err = db.Exec(`BEGIN TRANSACTION`)
    checkErr(err)
    _, err = db.Exec(`CREATE TEMPORARY TABLE operators_backup(MCCMNC_ID, Operator_Name, Country_ID)`)
    checkErr(err)
    _, err = db.Exec(`INSERT INTO operators_backup SELECT MCCMNC_ID, Operator_Name, Country_ID FROM operators`)
    checkErr(err)
    _, err = db.Exec(`DROP TABLE operators`)
    checkErr(err)
    _, err = db.Exec(`CREATE TABLE operators(MCCMNC_ID, Operator_Name, Country_ID)`)
    checkErr(err)
    _, err = db.Exec(`INSERT INTO operators SELECT MCCMNC_ID, Operator_Name, Country_ID FROM operators_backup;`)
    checkErr(err)
    _, err = db.Exec(`DROP TABLE operators_backup;`)
    checkErr(err)
    _, err = db.Exec(`COMMIT;`)
    checkErr(err)

    log.Println("initDB –\t\tInitializing countries table with temporary MCC table data...")
    _, err = db.Exec(`INSERT or IGNORE  INTO countries (Country_ID, name, MCC_ID) SELECT Country_ID, country, mcc FROM mytable`)
    checkErr(err)

    log.Println("initDB –\t\tLodaing config tables...")
    loadConfigTables(db)
    //==============================================//

    // //EXAMPLE SELECT CODE
    // log.Println("initDB –\t\tquerying mytable")
    // rows, err := db.Query("SELECT Operator_Name, Country_ID, MCCMNC_ID FROM mytable")
    // checkErr(err)
    // // rows, err := db.Query("SELECT COALESCE(mcc, '') || COALESCE(mnc, '') FROM mytable") //interesting way of concatting
    //
    // var Operator_Name string
    // var Country_ID string
    // var MCCMNC_ID string
    // for rows.Next() {
    //     rows.Scan(&Operator_Name, &Country_ID, &MCCMNC_ID)
    //     // log.Println("main –\t\t" + MCCMNC_ID + " | " + Operator_Name + " | " + Country_ID)
    // }
    // defer rows.Close()

    return db
}

// SELECT DISTINCT Config_ID, featuredLocationName FROM configurationMappings WHERE
// MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID = "in" )  //appTray for country=India without name

// SELECT DISTINCT appConfigs.Config_ID, originalName, configurationMappings.featuredLocationName FROM appConfigs
// JOIN     configurationMappings USING (Config_ID)
// WHERE Config_ID in (SELECT DISTINCT configurationMappings.Config_ID FROM configurationMappings WHERE
// MCCMNC_ID IN (SELECT MCCMNC_ID FROM operators WHERE Country_ID = "in" )) ;  //appTray for country=India


func similarConfigs(db *sql.DB, originalName string, Config_ID string) []string{
    var returnArrayOfConfig_IDs []string
    log.Println("similarConfigs –\tChecking appConfigs table for configs with originalName = " +originalName)
    rows, err := db.Query("SELECT Config_ID FROM appConfigs WHERE originalName = \"" + originalName+"\"")
    checkErr(err)

    var Config_ID2 string
    for rows.Next() {
        err = rows.Scan(&Config_ID2)
        checkErr(err)
        if(Config_ID != Config_ID2){
            log.Println("similarConfigs –\tAppName: "+originalName + " | Config_ID: " + Config_ID2)
            returnArrayOfConfig_IDs = append(returnArrayOfConfig_IDs, Config_ID2)
        }
    }

    rows.Close() //good habit to close

    return returnArrayOfConfig_IDs
}

func newAppConfig(db *sql.DB, Config_ID string, config_section string, featuredLocations string, originalName string, modifiableName string, iconURL string, homeURL string, rank string, versionNumber string) {
    log.Println("newAppConfig –\tInserting " + modifiableName + " " +config_section + " entries...")
    statement, _ := db.Prepare(`INSERT OR IGNORE INTO appConfigs (Config_ID, originalName, modifiableName , iconURL , homeURL , rank , versionNumber) VALUES (?, ?, ?, ?, ?, ?, ?)`)
    _, err := statement.Exec(Config_ID, originalName, modifiableName, iconURL, homeURL, rank, versionNumber)
    checkErr(err)
    var similarConfigs_IDs []string = similarConfigs(db, originalName, Config_ID)
    if(len(similarConfigs_IDs)>0){
        log.Println("newAppConfig –\tFound similarConfig:")
        log.Println(similarConfigs_IDs)
    }
    if(featuredLocations == "folder" || featuredLocations == "ALL") {
        execText := "INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT " + Config_ID + ", Country_ID FROM countries"
        _, err = db.Exec(execText)
        checkErr(err)
        execText = `INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "folder", Config_ID FROM appConfigs WHERE Config_ID = "`+Config_ID+`"`
        _, err = db.Exec(execText)
        checkErr(err)
    } else if(featuredLocations == "homescreen" || featuredLocations == "ALL") {
        execText := "INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT " + Config_ID + ", Country_ID FROM countries"
        _, err = db.Exec(execText)
        checkErr(err)
        execText = `INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "homescreen", Config_ID FROM appConfigs WHERE Config_ID = "`+Config_ID+`"`
        _, err = db.Exec(execText)
        checkErr(err)
    } else if(featuredLocations == "max" || featuredLocations == "ALL"){
        execText := "INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT " + Config_ID + ", Country_ID FROM countries"
        _, err = db.Exec(execText)
        checkErr(err)
        execText = `INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "max", Config_ID FROM appConfigs WHERE Config_ID = "`+Config_ID+`"`
        _, err = db.Exec(execText)
        checkErr(err)
    } else if(featuredLocations == "maxGo" || featuredLocations == "ALL"){
        execText := "INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT " + Config_ID + ", Country_ID FROM countries"
        _, err = db.Exec(execText)
        checkErr(err)
        execText = `INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "`+Config_ID+`"`
        _, err = db.Exec(execText)
        checkErr(err)
    } else if(strings.Contains(featuredLocations, ",")) {
        locations := strings.Split(featuredLocations, ",")
        for _, location := range locations {
            if location != ""{
                execText := "INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT " + Config_ID + ", Country_ID FROM countries"
                _, err = db.Exec(execText)
                checkErr(err)
                execText = `INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "`+location+`", Config_ID FROM appConfigs WHERE Config_ID = "`+Config_ID+`"`
                _, err = db.Exec(execText)
                checkErr(err)
            }
        }
    }
}


func loadConfigTables(db *sql.DB) {
    log.Println("loadConfigTables –\tInitializing [DEFAULT] appConfigs+configurationMappings tables...")
    // statement, _ := db.Prepare(`INSERT INTO appConfigs (originalName, modifiableName , iconURL , homeURL , rank , versionNumber) VALUES (?, ?, ?, ?, ?, ?)`)
    // =============================== (Config_ID* = 1) INSTAGRAM [DEFAULT] ==================================//
    newAppConfig(db, "1", "[DEFAULT]", "ALL", "instagram", "Instagram", "ultra_apps/instagram_ultra.png", "https://www.instagram.com/?utm_source=samsung_max_sd", "1", "3.1")

    newAppConfig(db, "2", "[DEFAULT]", "ALL", "cricbuzz", "Cricbuzz", "ultra_apps/cricbuzz_ultra.png","http://m.cricbuzz.com", "2" , "3.1")

    newAppConfig(db, "3", "[DEFAULT]", "ALL", "wikipedia", "Wikipedia", "ultra_apps/ic_wikipedia_ultra.png", "https://www.wikipedia.org", "3", "3.1")

    newAppConfig(db, "4", "[Global and Preloaded]", "ALL", "facebook", "Facebook", "ultra_apps/facebook_ultra_color.png", "https://m.facebook.com/?ref=s_max_bookmark", "4" , "3.1")
    // // =============================== (Config_ID* = 2) vKontakte [DEFAULT] ==================================//
    // log.Println("loadConfigTables –\tInserting vKontakte [DEFAULT] entries...")
    //
    // _, err := statement.Exec("vk", "vKontakte", "ultra_apps/vkontakte_ultra.png","https://vk.com", "4" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 2, Country_ID FROM countries WHERE Country_ID != "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "folder", Config_ID FROM appConfigs WHERE Config_ID = "2"`)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 2, Country_ID FROM countries WHERE Country_ID != "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "homescreen", Config_ID FROM appConfigs WHERE Config_ID = "2"`)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 2, Country_ID FROM countries WHERE Country_ID != "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "max", Config_ID FROM appConfigs WHERE Config_ID = "2"`)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 2, Country_ID FROM countries WHERE Country_ID != "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "2"`)
    // checkErr(err)
    //
    // // =============================== (Config_ID* = 3) Cricbuzz [DEFAULT] ==================================//
    // log.Println("loadConfigTables –\tInserting Cricbuzz [DEFAULT] entries...")
    // _, err = statement.Exec("cricbuzz", "Cricbuzz", "ultra_apps/cricbuzz_ultra.png","http://m.cricbuzz.com", "5" , "3.1")
    // checkErr(err)
    // //folder, homescreen, max
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 3, Country_ID FROM countries`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "folder", Config_ID FROM appConfigs WHERE Config_ID = "3"`)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 3, Country_ID FROM countries`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "homescreen", Config_ID FROM appConfigs WHERE Config_ID = "3"`)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 3, Country_ID FROM countries`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "max", Config_ID FROM appConfigs WHERE Config_ID = "3"`)
    // checkErr(err)
    //
    //
    // // =============================== (Config_ID* = 4) Wikipedia [DEFAULT] ==================================//
    // newAppConfig(db, "4", "[DEFAULT]", "ALL", "wikipedia", "Wikipedia", "ultra_apps/ic_wikipedia_ultra.png", "https://www.wikipedia.org", "7", "3.1")
    //
    // // =============================== (Config_ID* = 5) Facebook [Global and Preloaded] ==================================//
    // log.Println("loadConfigTables –\tInserting Facebook [Global and Preloaded] entries...")
    // newAppConfig(db, "5", "[Global and Preloaded]", "folder,homescreen,max", "facebook", "Facebook", "ultra_apps/facebook_ultra_color.png", "https://m.facebook.com/?ref=s_max_bookmark", "1" , "3.1")
    // // =============================== (Config_ID = 1) Instagram [Global and Preloaded] ==================================//
    // log.Println("loadConfigTables –\tInserting Instagram [Global and Preloaded] entries...")
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 1, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 1, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 1, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    // // =============================== (Config_ID = 2) vKontakte [Global and Preloaded] ==================================//
    // log.Println("loadConfigTables –\tInserting vKontakte [Global and Preloaded] entries...")
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 2, Country_ID FROM countries WHERE Country_ID != "in"`)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 2, Country_ID FROM countries WHERE Country_ID != "in"`)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 2, Country_ID FROM countries WHERE Country_ID != "in"`)
    // checkErr(err)
    //
    //
    // // =============================== (Config_ID = 3) Cricbuzz [Global and Preloaded] ==================================//
    // log.Println("loadConfigTables –\tInserting Cricbuzz [Global and Preloaded] entries...")
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 3, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 3, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 3, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    // // =============================== (Config_ID = 4) Wikipedia [Global and Preloaded] ==================================//
    // log.Println("loadConfigTables –\tInserting Wikipedia [Global and Preloaded] entries...")
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 4, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 4, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 4, Country_ID FROM countries`)
    // checkErr(err)
    //
    //
    // // =============================== (Config_ID = 5) Facebook [global_and_preloaded_with_freebasic] ==================================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 5, MCCMNC_ID FROM operators WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 5, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 5, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    //
    // // ========================== (Config_ID = 1) Instagram [global_and_preloaded_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 1, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 1, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 1, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 2) vKontakte [global_and_preloaded_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 2, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID != "in"`)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 2, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID != "in"`)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 2, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID != "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 3) Cricbuzz [global_and_preloaded_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 3, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 3, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 3, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 6) Free Basics [global_and_preloaded_with_freebasic] =============================//
    // log.Println("loadConfigTables –\tInserting Free Basics [global_and_preloaded_with_freebasic] entries...")
    // _, err = statement.Exec("freebasics", "Free Basics", "ultra_apps/ic_free_basics.png","https://freebasics.com/?ref=s_max_bookmark", "6" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 6, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "folder", Config_ID FROM appConfigs WHERE Config_ID = "6"`)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 6, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "homescreen", Config_ID FROM appConfigs WHERE Config_ID = "6"`)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 6, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "max", Config_ID FROM appConfigs WHERE Config_ID = "6"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 4) Wikipedia [global_and_preloaded_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 4, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 4, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 4, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 7) Facebook [android_go] =============================//
    // log.Println("loadConfigTables –\tInserting Facebook [android_go] entry...")
    // _, err = statement.Exec("facebook", "Facebook", "ultra_apps/facebook_ultra_color.png","https://m.facebook.com/?ref=s_max_bookmark", "1" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 7, Country_ID FROM countries`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "7"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 8) Instagram [android_go] =============================//
    // log.Println("loadConfigTables –\tInserting Instagram [android_go] entry...")
    // _, err = statement.Exec("instagram", "Instagram", "ultra_apps/instagram_ultra.png","https://www.instagram.com/?utm_source=samsung_max_sd", "2" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 8, Country_ID FROM countries`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "8"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 9) Twitter [android_go] =============================//
    // log.Println("loadConfigTables –\tInserting Twitter [android_go] entry...")
    // _, err = statement.Exec("twitter", "Twitter", "ultra_apps/twitter_ultra.png","https://mobile.twitter.com", "3" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 9, Country_ID FROM countries`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "9"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 10) vKontakte [android_go] =============================//
    // log.Println("loadConfigTables –\tInserting vKontakte [android_go] entry...")
    // _, err = statement.Exec("vk", "vKontakte", "ultra_apps/vkontakte_ultra.png","https://vk.com", "4" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 10, Country_ID FROM countries WHERE Country_ID!="in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "10"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 11) Wikipedia [android_go] =============================//
    // log.Println("loadConfigTables –\tInserting Twitter [android_go] entry...")
    // _, err = statement.Exec("wikipedia", "Wikipedia", "ultra_apps/ic_wikipedia_ultra.png","https://www.wikipedia.org", "7" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 11, Country_ID FROM countries`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "11"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 12) Worldreader [android_go] =============================//
    // log.Println("loadConfigTables –\tInserting worldreader [android_go] entry...")
    // _, err = statement.Exec("worldreader", "Worldreader", "ultra_apps/ic_worldreader_ultra.png","https://www.worldreader.org", "25" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 12, Country_ID FROM countries`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "12"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 7) Facebook [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 7, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 8) Instagram [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 8, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 9) Twitter [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 9, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) `)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 10) vKontakte [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 10, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics" )) AND Country_ID != "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 7) Freebasics [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 7, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics"))`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 6) Freebasics [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 6, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics"))`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 11) Wikipedia [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 11, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics"))`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 12) Worldreader [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 12, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics"))`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 12) Worldreader [android_go_with_freebasic] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 12, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics"))`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 13) DailyHunt [android_go_india] =============================//
    // log.Println("loadConfigTables –\tInserting DailyHunt [android_go_india] entry...")
    // _, err = statement.Exec("dailyhunt", "Dailyhunt", "ultra_apps/ic_dailyhunt_lite.png","https://samsung.dailyhunt.in/news/india/english?utm_source=Samsung&utm_medium=Android%20Go&utm_campaign=Phase%201&mode=pwa&s=Samsung&ss=AndroidGo", "0" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 13, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "13"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 7) Facebook [android_go_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 7, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 8) Instagram [android_go_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 8, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 9) Twitter [android_go_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 9, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 11) Wikipedia [android_go_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 11, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 14) Nitrostreet [android_go_india] =============================//
    // log.Println("loadConfigTables –\tInserting Nitrostreet [android_go_india] entry...")
    // _, err = statement.Exec("nitrostreet", "Nitro StreetRun 2", "ultra_apps/ic_nitrostreet_ultra.png","http://play.ludigames.com/games/nitroStreetRun2Free/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH72", "20" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 14, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "14"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 15) PuzzlePets [android_go_india] =============================//
    // log.Println("loadConfigTables –\tInserting PuzzlePets [android_go_india] entry...")
    // _, err = statement.Exec("puzzlepets", "Puzzle Pets Pairs", "ultra_apps/ic_puzzlepets_ultra.png","http://play.ludigames.com/games/puzzlePetsPairsFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH72", "21" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 15, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "15"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 16) Paper Flight [android_go_india] =============================//
    // log.Println("loadConfigTables –\tInserting Paper Flight [android_go_india] entry...")
    // _, err = statement.Exec("paperflight", "Paper Flight", "ultra_apps/ic_paperflight_ultra.png","http://play.ludigames.com/games/paperFlightFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH72", "22" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 16, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "16"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 17) Ludibubbles [android_go_india] =============================//
    // log.Println("loadConfigTables –\tInserting Ludibubbles [android_go_india] entry...")
    // _, err = statement.Exec("ludibubbles", "Ludibubbles", "ultra_apps/ic_ludibubbles_ultra.png","http://play.ludigames.com/games/ludibubblesFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH72", "23" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 17, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "17"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID* = 18) Moregames [android_go_india] =============================//
    // log.Println("loadConfigTables –\tInserting Moregames [android_go_india] entry...")
    // _, err = statement.Exec("moregames", "More games", "ultra_apps/ic_more_games.png","http://play.ludigames.com/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH72", "24" , "3.1")
    // checkErr(err)
    //
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 18, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    // _, err = db.Exec(`INSERT OR IGNORE INTO featuredLocations (featuredLocationName, Config_ID) SELECT "maxGo", Config_ID FROM appConfigs WHERE Config_ID = "18"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 12) Worldreader [android_go_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, Country_ID) SELECT 12, Country_ID FROM countries WHERE Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 13) DailyHunt [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 13, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 7) Facebook [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 7, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 8) Instagram [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 8, MCCMNC_ID FROM operators WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 9) Twitter [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 9, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 6) Free Basics [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 6, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 11) Wikipedia [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 11, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 14) Nitrostreet [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 14, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 15) PuzzlePets [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 15, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 16) PaperFlight [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 16, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 17) Ludibubbles [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 17, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 18) Moregames [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 18, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
    //
    // // ========================== (Config_ID = 12) Worldreader [android_go_with_freebasic_india] =============================//
    // _, err = db.Exec(`INSERT OR IGNORE INTO configurationMappings (Config_ID, MCCMNC_ID) SELECT 12, MCCMNC_ID FROM operators  WHERE MCCMNC_ID in ( SELECT MCCMNC_ID from operatorGroups WHERE  (Operator_Group_Name = "freebasics" OR Operator_Group_Name = "viettel" OR Operator_Group_Name = "digicel-pa" OR Operator_Group_Name = "telcel" OR Operator_Group_Name = "tigo-co" OR Operator_Group_Name = "viva-bo" OR Operator_Group_Name = "mobifone" OR Operator_Group_Name = "freebasics")) AND Country_ID = "in"`)
    // checkErr(err)
}

func createTables(db *sql.DB) {
    log.Println( "createTables –\tCreating MobileCountryCodeDB temp table...")
    createMobileCountryCodeDB(db)


    log.Println("createTables –\tDropping countries table if exists...")
    _, err := db.Exec("DROP TABLE IF EXISTS countries")

    log.Println( "createTables –\tCreating countries table...")
    stmt, _ := db.Prepare("CREATE TABLE IF NOT EXISTS countries ( Country_ID TEXT PRIMARY KEY, name TEXT NOT NULL, MCC_ID INTEGER NOT NULL)")
    _, err = stmt.Exec()
    checkErr(err)

    log.Println("createTables –\tDropping operators table if exists...")
    _, err = db.Exec("DROP TABLE IF EXISTS operators")

    log.Println( "createTables –\tCreating operators table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS operators ( MCCMNC_ID TEXT PRIMARY KEY, Operator_Name TEXT, Operator_Group_Name TEXT, Country_ID TEXT,  FOREIGN KEY(Country_ID) REFERENCES countries(Country_ID) )")
    _, err = stmt.Exec()
    checkErr(err)

    log.Println("createTables –\tDropping operatorGroups table if exists...")
    _, err = db.Exec("DROP TABLE IF EXISTS operatorGroups")

    log.Println( "createTables –\tCreating operatorGroups table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS operatorGroups ( id INTEGER PRIMARY KEY AUTOINCREMENT, Operator_Group_Name TEXT, MCCMNC_ID TEXT, FOREIGN KEY(MCCMNC_ID) REFERENCES operators(MCCMNC_ID) )")
    _, err = stmt.Exec()
    checkErr(err)

    log.Println("createTables –\tDropping appConfigs table if exists...")
    _, err = db.Exec("DROP TABLE IF EXISTS appConfigs")

    log.Println( "createTables –\tCreating appConfigs table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS appConfigs ( Config_ID INTEGER PRIMARY KEY AUTOINCREMENT, originalName TEXT, modifiableName TEXT, iconURL TEXT, homeURL TEXT, rank INTEGER, category TEXT, versionNumber FLOAT, FOREIGN KEY(versionNumber) REFERENCES versions(versionNumber))")
    _, err = stmt.Exec()
    checkErr(err)

    log.Println("createTables –\tDropping configurationMappings table if exists...")
    _, err = db.Exec("DROP TABLE IF EXISTS configurationMappings")

    log.Println( "createTables –\tCreating configurationMappings table...")
    stmt, _ = db.Prepare("CREATE TABLE IF NOT EXISTS configurationMappings ( id INTEGER PRIMARY KEY AUTOINCREMENT, Config_ID INTEGER, MCCMNC_ID TEXT, Country_ID TEXT,  FOREIGN KEY(Config_ID) REFERENCES appConfigs(Config_ID), FOREIGN KEY(MCCMNC_ID) REFERENCES operators(MCCMNC_ID), FOREIGN KEY(Country_ID) REFERENCES countries(Country_ID))")
    _, err = stmt.Exec()
    checkErr(err)

    log.Println("createTables –\tDropping featuredLocations table if exists...")
    _, err = db.Exec("DROP TABLE IF EXISTS featuredLocations")
    log.Println( "createTables –\tCreating featuredLocations table...")
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS featuredLocations (id INTEGER PRIMARY KEY AUTOINCREMENT, featuredLocationName TEXT, Config_ID INTEGER, FOREIGN KEY(Config_ID) REFERENCES appConfigs(Config_ID), UNIQUE(featuredLocationName, Config_ID))")
    checkErr(err)


    log.Println("createTables –\tDropping featureMappings table if exists...")
    _, err = db.Exec("DROP TABLE IF EXISTS featureMappings")
    log.Println( "createTables –\tCreating featureMappings table...")
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS featureMappings (id INTEGER PRIMARY KEY AUTOINCREMENT, featureType TEXT, featureName TEXT, Config_ID INTERGER, FOREIGN KEY(Config_ID) REFERENCES appConfigs(Config_ID))")
    checkErr(err)


    log.Println("createTables –\tDropping users table if exists...")
    _, err = db.Exec("DROP TABLE IF EXISTS users")
    log.Println( "createTables –\tCreating users table...")
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS users (userID INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, password TEXT)")
    checkErr(err)

    log.Println("createTables –\tDropping userSessions table if exists...")
    _, err = db.Exec("DROP TABLE IF EXISTS userSessions")
    log.Println( "createTables –\tCreating userSessions table...")
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS userSessions (sessionKey TEXT PRIMARY KEY, userID INTEGER, FOREIGN KEY(userID) REFERENCES users(userID) )")
    checkErr(err)
}


func createMobileCountryCodeDB(db *sql.DB){
    stmt, _ := db.Prepare(`DROP TABLE IF EXISTS mytable`)
    _, err := stmt.Exec()
    checkErr(err)
    stmt, _ = db.Prepare(`DROP TABLE IF EXISTS mytable2`)
    _, err = stmt.Exec()
    checkErr(err)

    stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS mytable(
        id            INTEGER  AUTO_INCREMENT  PRIMARY KEY
        ,Operator_Name      VARCHAR(72)
        ,country      VARCHAR(33) NOT NULL
        ,mcc          INTEGER NOT NULL
        ,Country_ID          VARCHAR(3) NOT NULL
        ,country_code INTEGER
        ,mnc          VARCHAR(3) NOT NULL
        ,MCCMNC_ID    VARCHAR(6)
        )`)
    _, err = stmt.Exec()
    checkErr(err)
    stmt, _ = db.Prepare(`CREATE TABLE IF NOT EXISTS mytable2(
        MCCMNC_ID            INTEGER  PRIMARY KEY
        ,Operator_Group_Name      VARCHAR(72)
        )`)
    _, err = stmt.Exec()
    checkErr(err)

    //OUR OPERATORS
    freebasics:=[...]string{"338050","342050","344930","35850","36269","37001","374130","376050","40409","40418","40436","40450","40452","40467","40485","41004","41006","41401","41601","41603","41805","41820","41840","41882","42800","42891","42898","42899","45603","45605","45606","47001","47002","47202","51001","51010","51021","51401","51501","51502","51503","51505","52004","52099","53703","54101","60203","60303","60802","60910","61105","61205","61402","61502","61503","61602","61603","61807","62002","62003","62006","62120","62160","62203","62401","62501","62502","62803","62901","63086","63089","63104","63310","63401","63406","63513","63514","63903","64002","64004","64005", "64301","64501","64502","64503","64602","65001","65010","70402","70602","732001","732130","74603"}
    att_mx:=[...]string{"33401","334010","334050","334090"}
    stmt, err = db.Prepare(`INSERT OR IGNORE INTO mytable2 (Operator_Group_Name, MCCMNC_ID) VALUES (?,?)`)
    checkErr(err)
    _, err = stmt.Exec("alegro-ec", "74002")
    checkErr(err)
    for _, operator := range att_mx {
        _, err = stmt.Exec("att_mx", operator)
    }
    _, err = stmt.Exec("beeline", "45207")
    _, err = stmt.Exec("bitel-pe", "71615")
    _, err = stmt.Exec("claro-co", "732101")
    _, err = stmt.Exec("claro-ec", "74001")
    _, err = stmt.Exec("claro-pa", "71403")
    _, err = stmt.Exec("claro-pe", "71610")
    _, err = stmt.Exec("claro-pr", "330110")
    _, err = stmt.Exec("cwp-pa", "71401")
    _, err = stmt.Exec("digicel-pa", "71404")
    _, err = stmt.Exec("entel-pe", "71607")
    _, err = stmt.Exec("entel-pe", "71617")
    for _, operator := range freebasics {
        _, err = stmt.Exec("freebasics", operator)
    }
    _, err = stmt.Exec("mobifone", "45201")
    _, err = stmt.Exec("movistar-ec", "74000")
    _, err = stmt.Exec("movistar-gt", "70403")
    _, err = stmt.Exec("movistar-pa", "70403")
    _, err = stmt.Exec("movistar-pa", "714020")
    _, err = stmt.Exec("movistar-pe", "71606")
    _, err = stmt.Exec("ncell-np", "42902")
    _, err = stmt.Exec("sfone", "45203")
    _, err = stmt.Exec("telcel", "334020")
    _, err = stmt.Exec("tigo-co", "732103")
    _, err = stmt.Exec("tigo-co", "732111")
    _, err = stmt.Exec("uiautomator", "00103")
    _, err = stmt.Exec("viettel", "45204")
    _, err = stmt.Exec("viettel", "45206")
    _, err = stmt.Exec("viettel", "45208")
    _, err = stmt.Exec("vietnamobile", "45205")
    _, err = stmt.Exec("vinaphone", "45202")
    _, err = stmt.Exec("viva-bo", "73601")


    sqlInserts := `INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('A-Mobile','Abkhazia',289,'ge',7,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('A-Mobile','Abkhazia',289,'ge',7,'68');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aquafon','Abkhazia',289,'ge',7,'67');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Afghan Telecom Corp. (AT)','Afghanistan',412,'af',93,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Afghan Telecom Corp. (AT)','Afghanistan',412,'af',93,'80');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Afghan Wireless/AWCC','Afghanistan',412,'af',93,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Areeba/MTN','Afghanistan',412,'af',93,'40');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Etisalat','Afghanistan',412,'af',93,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Etisalat','Afghanistan',412,'af',93,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Roshan/TDCA','Afghanistan',412,'af',93,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('WaselTelecom (WT)','Afghanistan',412,'af',93,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AMC/Cosmote','Albania',276,'al',355,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Eagle Mobile','Albania',276,'al',355,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PLUS Communication Sh.a','Albania',276,'al',355,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Albania',276,'al',355,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ATM Mobils','Algeria',603,'dz',213,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orascom / DJEZZY','Algeria',603,'dz',213,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Oreedo/Wataniya / Nedjma','Algeria',603,'dz',213,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Blue Sky Communications','American Samoa',544,'as',684,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobiland','Andorra',213,'ad',376,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MoviCel','Angola',631,'ao',244,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unitel','Angola',631,'ao',244,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable and Wireless','Anguilla',365,'ai',1264,'840');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicell / Wireless Vent. Ltd','Anguilla',365,'ai',1264,'010');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('APUA PCS','Antigua and Barbuda',344,'ag',1268,'030');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('C & W','Antigua and Barbuda',344,'ag',1268,'920');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiCel/Cing. Wireless','Antigua and Barbuda',344,'ag',1268,'930');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/ CTI/AMX','Argentina Republic',722,'ar',54,'310');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/ CTI/AMX','Argentina Republic',722,'ar',54,'330');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/ CTI/AMX','Argentina Republic',722,'ar',54,'320');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Compania De Radiocomunicaciones Moviles SA','Argentina Republic',722,'ar',54,'010');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar/Telefonica','Argentina Republic',722,'ar',54,'070');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel','Argentina Republic',722,'ar',54,'020');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecom Personal S.A.','Argentina Republic',722,'ar',54,'341');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecom Personal S.A.','Argentina Republic',722,'ar',54,'340');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ArmenTel/Beeline','Armenia',283,'am',374,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Karabakh Telecom','Armenia',283,'am',374,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Armenia',283,'am',374,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivacell','Armenia',283,'am',374,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Aruba',363,'aw',297,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Aruba',363,'aw',297,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Setar GSM','Aruba',363,'aw',297,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AAPT Ltd.','Australia',505,'au',61,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Advanced Comm Tech Pty.','Australia',505,'au',61,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airnet Commercial Australia Ltd..','Australia',505,'au',61,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Department of Defense','Australia',505,'au',61,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Dialogue Communications Pty Ltd','Australia',505,'au',61,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G Ltd.','Australia',505,'au',61,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G Ltd.','Australia',505,'au',61,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Localstar Holding Pty. Ltd','Australia',505,'au',61,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile Pty Ltd','Australia',505,'au',61,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Railcorp/Vodafone','Australia',505,'au',61,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Railcorp/Vodafone','Australia',505,'au',61,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Railcorp/Vodafone','Australia',505,'au',61,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Singtel Optus','Australia',505,'au',61,'90');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Singtel Optus','Australia',505,'au',61,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telstra Corp. Ltd.','Australia',505,'au',61,'71');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telstra Corp. Ltd.','Australia',505,'au',61,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telstra Corp. Ltd.','Australia',505,'au',61,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telstra Corp. Ltd.','Australia',505,'au',61,'72');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('The Ozitel Network Pty.','Australia',505,'au',61,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Victorian Rail Track Corp. (VicTrack)','Australia',505,'au',61,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Australia',505,'au',61,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Australia',505,'au',61,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('A1 MobilKom','Austria',232,'at',43,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('A1 MobilKom','Austria',232,'at',43,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('A1 MobilKom','Austria',232,'at',43,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('A1 MobilKom','Austria',232,'at',43,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile/Telering','Austria',232,'at',43,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G','Austria',232,'at',43,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G','Austria',232,'at',43,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('3/Orange/One Connect','Austria',232,'at',43,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('3/Orange/One Connect','Austria',232,'at',43,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('3/Orange/One Connect','Austria',232,'at',43,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Spusu/Mass Response','Austria',232,'at',43,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile/Telering','Austria',232,'at',43,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile/Telering','Austria',232,'at',43,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile/Telering','Austria',232,'at',43,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2','Austria',232,'at',43,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('A1 MobilKom','Austria',232,'at',43,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UPC Austria','Austria',232,'at',43,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Azercell Telekom B.M.','Azerbaijan',400,'az',994,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Azerfon.','Azerbaijan',400,'az',994,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Caspian American Telecommunications LLC (CATEL)','Azerbaijan',400,'az',994,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('J.V. Bakcell GSM 2000','Azerbaijan',400,'az',994,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bahamas Telco. Comp.','Bahamas',364,'bs',1242,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bahamas Telco. Comp.','Bahamas',364,'bs',1242,'390');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bahamas Telco. Comp.','Bahamas',364,'bs',1242,'39');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smart Communications','Bahamas',364,'bs',1242,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Batelco','Bahrain',426,'bh',973,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ZAIN/Vodafone','Bahrain',426,'bh',973,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VIVA','Bahrain',426,'bh',973,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Robi/Aktel','Bangladesh',470,'bd',880,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Citycell','Bangladesh',470,'bd',880,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Citycell','Bangladesh',470,'bd',880,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GrameenPhone','Bangladesh',470,'bd',880,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orascom/Banglalink','Bangladesh',470,'bd',880,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TeleTalk','Bangladesh',470,'bd',880,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Warid','Bangladesh',470,'bd',880,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('LIME','Barbados',342,'bb',1246,'600');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cingular Wireless','Barbados',342,'bb',1246,'810');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Barbados',342,'bb',1246,'750');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Barbados',342,'bb',1246,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sunbeach','Barbados',342,'bb',1246,'820');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BelCel JV','Belarus',257,'by',375,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BeST','Belarus',257,'by',375,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobile Digital Communications','Belarus',257,'by',375,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTS','Belarus',257,'by',375,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Base/KPN','Belgium',206,'be',32,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Belgacom/Proximus','Belgium',206,'be',32,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile Belgium','Belgium',206,'be',32,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobistar/Orange','Belgium',206,'be',32,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SNCT/NMBS','Belgium',206,'be',32,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenet BidCo NV','Belgium',206,'be',32,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiCell','Belize',702,'bz',501,'67');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('International Telco (INTELCO)','Belize',702,'bz',501,'68');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bell Benin/BBCOM','Benin',616,'bj',229,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Etisalat/MOOV','Benin',616,'bj',229,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GloMobile','Benin',616,'bj',229,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Libercom','Benin',616,'bj',229,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Spacetel','Benin',616,'bj',229,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bermuda Digital Communications Ltd (BDC)','Bermuda',350,'bm',1441,'000');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CellOne Ltd','Bermuda',350,'bm',1441,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiCel / Cingular','Bermuda',350,'bm',1441,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('M3 Wireless Ltd','Bermuda',350,'bm',1441,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecommunications (Bermuda & West Indies) Ltd (Digicel Bermuda)','Bermuda',350,'bm',1441,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('B-Mobile','Bhutan',402,'bt',975,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bhutan Telecom Ltd (BTL)','Bhutan',402,'bt',975,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TashiCell','Bhutan',402,'bt',975,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Entel Pcs','Bolivia',736,'bo',591,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Viva/Nuevatel','Bolivia',736,'bo',591,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tigo','Bolivia',736,'bo',591,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BH Mobile','Bosnia & Herzegov.',218,'ba',387,'90');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Eronet Mobile','Bosnia & Herzegov.',218,'ba',387,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('M-Tel','Bosnia & Herzegov.',218,'ba',387,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BeMOBILE','Botswana',652,'bw',267,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mascom Wireless (Pty) Ltd.','Botswana',652,'bw',267,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Botswana',652,'bw',267,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/Albra/America Movil','Brazil',724,'br',55,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/Albra/America Movil','Brazil',724,'br',55,'38');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/Albra/America Movil','Brazil',724,'br',55,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivo S.A./Telemig','Brazil',724,'br',55,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CTBC Celular SA (CTBC)','Brazil',724,'br',55,'33');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CTBC Celular SA (CTBC)','Brazil',724,'br',55,'32');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CTBC Celular SA (CTBC)','Brazil',724,'br',55,'34');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIM','Brazil',724,'br',55,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel (Telet)','Brazil',724,'br',55,'39');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel (Telet)','Brazil',724,'br',55,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Oi (TNL PCS / Oi)','Brazil',724,'br',55,'31');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Brazil Telcom','Brazil',724,'br',55,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Amazonia Celular S/A','Brazil',724,'br',55,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Oi (TNL PCS / Oi)','Brazil',724,'br',55,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PORTO SEGURO TELECOMUNICACOES','Brazil',724,'br',55,'54');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sercontel Cel','Brazil',724,'br',55,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CTBC/Triangulo','Brazil',724,'br',55,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivo S.A./Telemig','Brazil',724,'br',55,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIM','Brazil',724,'br',55,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIM','Brazil',724,'br',55,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIM','Brazil',724,'br',55,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unicel do Brasil Telecomunicacoes Ltda','Brazil',724,'br',55,'37');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivo S.A./Telemig','Brazil',724,'br',55,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivo S.A./Telemig','Brazil',724,'br',55,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivo S.A./Telemig','Brazil',724,'br',55,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivo S.A./Telemig','Brazil',724,'br',55,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Caribbean Cellular','British Virgin Islands',348,'vg',284,'570');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','British Virgin Islands',348,'vg',284,'770');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('LIME','British Virgin Islands',348,'vg',284,'170');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('b-mobile','Brunei Darussalam',528,'bn',673,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Datastream (DTSCom)','Brunei Darussalam',528,'bn',673,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telekom Brunei Bhd (TelBru)','Brunei Darussalam',528,'bn',673,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BTC Mobile EOOD (vivatel)','Bulgaria',284,'bg',359,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BTC Mobile EOOD (vivatel)','Bulgaria',284,'bg',359,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor/Cosmo/Globul','Bulgaria',284,'bg',359,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MobilTel AD','Bulgaria',284,'bg',359,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TeleCel','Burkina Faso',613,'bf',226,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TeleMob-OnaTel','Burkina Faso',613,'bf',226,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/ZAIN/CelTel','Burkina Faso',613,'bf',226,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Africel / Safaris','Burundi',642,'bi',257,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lumitel/Viettel','Burundi',642,'bi',257,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Onatel / Telecel','Burundi',642,'bi',257,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smart Mobile / LACELL','Burundi',642,'bi',257,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Spacetel / Econet / Leo','Burundi',642,'bi',257,'82');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Spacetel / Econet / Leo','Burundi',642,'bi',257,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cambodia Advance Communications Co. Ltd (CADCOMMS)','Cambodia',456,'kh',855,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smart Mobile','Cambodia',456,'kh',855,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Metfone','Cambodia',456,'kh',855,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MFone/Camshin/Cellcard','Cambodia',456,'kh',855,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobitel/Cam GSM','Cambodia',456,'kh',855,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('QB/Cambodia Adv. Comms.','Cambodia',456,'kh',855,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smart Mobile','Cambodia',456,'kh',855,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smart Mobile','Cambodia',456,'kh',855,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sotelco/Beeline','Cambodia',456,'kh',855,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN','Cameroon',624,'cm',237,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel','Cameroon',624,'cm',237,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Cameroon',624,'cm',237,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BC Tel Mobility','Canada',302,'ca',1,'652');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bell Aliant','Canada',302,'ca',1,'630');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bell Mobility','Canada',302,'ca',1,'610');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bell Mobility','Canada',302,'ca',1,'651');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CityWest Mobility','Canada',302,'ca',1,'670');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Clearnet','Canada',302,'ca',1,'361');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Clearnet','Canada',302,'ca',1,'360');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DMTS Mobility','Canada',302,'ca',1,'380');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Globalstar Canada','Canada',302,'ca',1,'710');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Latitude Wireless','Canada',302,'ca',1,'640');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('FIDO (Rogers AT&T/ Microcell)','Canada',302,'ca',1,'370');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('mobilicity','Canada',302,'ca',1,'320');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MT&T Mobility','Canada',302,'ca',1,'702');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTS Mobility','Canada',302,'ca',1,'655');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTS Mobility','Canada',302,'ca',1,'660');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NB Tel Mobility','Canada',302,'ca',1,'701');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('New Tel Mobility','Canada',302,'ca',1,'703');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Public Mobile','Canada',302,'ca',1,'760');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Quebectel Mobility','Canada',302,'ca',1,'657');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Rogers AT&T Wireless','Canada',302,'ca',1,'720');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sask Tel Mobility','Canada',302,'ca',1,'654');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sask Tel Mobility','Canada',302,'ca',1,'780');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sask Tel Mobility','Canada',302,'ca',1,'680');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tbay Mobility','Canada',302,'ca',1,'656');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telus Mobility','Canada',302,'ca',1,'653');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telus Mobility','Canada',302,'ca',1,'220');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Videotron','Canada',302,'ca',1,'500');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('WIND','Canada',302,'ca',1,'490');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CV Movel','Cape Verde',625,'cv',238,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T+ Telecom','Cape Verde',625,'cv',238,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel Cayman Ltd','Cayman Islands',346,'ky',1345,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel Ltd.','Cayman Islands',346,'ky',1345,'006');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('LIME / Cable & Wirel.','Cayman Islands',346,'ky',1345,'140');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Centrafr. Telecom+','Central African Rep.',623,'cf',236,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nationlink','Central African Rep.',623,'cf',236,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Celca','Central African Rep.',623,'cf',236,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecel Centraf.','Central African Rep.',623,'cf',236,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Salam/Sotel','Chad',622,'td',235,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tchad Mobile','Chad',622,'td',235,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tigo/Milicom/Tchad Mobile','Chad',622,'td',235,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/ZAIN/Celtel','Chad',622,'td',235,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Blue Two Chile SA','Chile',730,'cl',56,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Celupago SA','Chile',730,'cl',56,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cibeles Telecom SA','Chile',730,'cl',56,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro','Chile',730,'cl',56,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Entel Telefonia','Chile',730,'cl',56,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Entel Telefonia Mov','Chile',730,'cl',56,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Netline Telefonica Movil Ltda','Chile',730,'cl',56,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel SA','Chile',730,'cl',56,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel SA','Chile',730,'cl',56,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel SA','Chile',730,'cl',56,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sociedad Falabella Movil SPA','Chile',730,'cl',56,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TELEFONICA','Chile',730,'cl',56,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TELEFONICA','Chile',730,'cl',56,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telestar Movil SA','Chile',730,'cl',56,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TESAM SA','Chile',730,'cl',56,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tribe Mobile SPA','Chile',730,'cl',56,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VTR Banda Ancha SA','Chile',730,'cl',56,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Mobile GSM','China',460,'cn',86,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Mobile GSM','China',460,'cn',86,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Mobile GSM','China',460,'cn',86,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Space Mobile Satellite Telecommunications Co. Ltd (China Spacecom)','China',460,'cn',86,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Telecom','China',460,'cn',86,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Telecom','China',460,'cn',86,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Unicom','China',460,'cn',86,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Unicom','China',460,'cn',86,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Avantel SAS','Colombia',732,'co',57,'130');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Colombia',732,'co',57,'102');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIGO/Colombia Movil','Colombia',732,'co',57,'103');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIGO/Colombia Movil','Colombia',732,'co',57,'001');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comcel S.A. Occel S.A./Celcaribe','Colombia',732,'co',57,'101');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Edatel S.A.','Colombia',732,'co',57,'002');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('eTb','Colombia',732,'co',57,'187');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Colombia',732,'co',57,'123');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIGO/Colombia Movil','Colombia',732,'co',57,'111');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UNE EPM Telecomunicaciones SA ESP','Colombia',732,'co',57,'142');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UNE EPM Telecomunicaciones SA ESP','Colombia',732,'co',57,'020');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Virgin Mobile Colombia SAS','Colombia',732,'co',57,'154');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HURI - SNPT','Comoros',654,'km',269,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Africell','Congo, Dem. Rep.',630,'cd',243,'90');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange RDC sarl','Congo, Dem. Rep.',630,'cd',243,'86');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SuperCell','Congo, Dem. Rep.',630,'cd',243,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIGO/Oasis','Congo, Dem. Rep.',630,'cd',243,'89');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodacom','Congo, Dem. Rep.',630,'cd',243,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Yozma Timeturns sprl (YTT)','Congo, Dem. Rep.',630,'cd',243,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/ZAIN','Congo, Dem. Rep.',630,'cd',243,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel SA','Congo, Republic',629,'cg',242,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Azur SA (ETC)','Congo, Republic',629,'cg',242,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Libertis','Congo, Republic',629,'cg',242,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Warid','Congo, Republic',629,'cg',242,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecom Cook Islands','Cook Islands',548,'ck',682,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro','Costa Rica',712,'cr',506,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ICE','Costa Rica',712,'cr',506,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ICE','Costa Rica',712,'cr',506,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Costa Rica',712,'cr',506,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Virtualis','Costa Rica',712,'cr',506,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile/Cronet','Croatia',219,'hr',385,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2','Croatia',219,'hr',385,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VIPnet d.o.o.','Croatia',219,'hr',385,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('C-COM','Cuba',368,'cu',53,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('EOCG Wireless NV','Curacao',362,'cw',599,'95');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Polycom N.V./ Digicel','Curacao',362,'cw',599,'69');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Areeba','Cyprus',280,'cy',357,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PrimeTel PLC','Cyprus',280,'cy',357,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone/CyTa','Cyprus',280,'cy',357,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Compatel s.r.o.','Czech Rep.',230,'cz',420,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O2','Czech Rep.',230,'cz',420,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile / RadioMobil','Czech Rep.',230,'cz',420,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Travel Telekommunikation s.r.o.','Czech Rep.',230,'cz',420,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ufone','Czech Rep.',230,'cz',420,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Czech Rep.',230,'cz',420,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Czech Rep.',230,'cz',420,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ApS KBUS','Denmark',238,'dk',45,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Banedanmark','Denmark',238,'dk',45,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CoolTEL ApS','Denmark',238,'dk',45,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G','Denmark',238,'dk',45,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile Ltd','Denmark',238,'dk',45,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mach Connectivity ApS','Denmark',238,'dk',45,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mundio Mobile','Denmark',238,'dk',45,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NextGen Mobile Ltd (CardBoardFish)','Denmark',238,'dk',45,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Denmark','Denmark',238,'dk',45,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Denmark','Denmark',238,'dk',45,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor/Sonofon','Denmark',238,'dk',45,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor/Sonofon','Denmark',238,'dk',45,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telia','Denmark',238,'dk',45,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telia','Denmark',238,'dk',45,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Djibouti Telecom SA (Evatis)','Djibouti',638,'dj',253,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('C & W','Dominica',366,'dm',1767,'110');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cingular Wireless/Digicel','Dominica',366,'dm',1767,'020');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wireless Ventures (Dominica) Ltd (Digicel Dominica)','Dominica',366,'dm',1767,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro','Dominican Republic',370,'do',1809,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Dominican Republic',370,'do',1809,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TRIcom','Dominican Republic',370,'do',1809,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Trilogy Dominicana S. A.','Dominican Republic',370,'do',1809,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Alegro/Telcsa','Ecuador',740,'ec',593,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MOVISTAR/OteCel','Ecuador',740,'ec',593,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/Port','Ecuador',740,'ec',593,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Mobinil','Egypt',602,'eg',20,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ETISALAT','Egypt',602,'eg',20,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone/Mirsfone','Egypt',602,'eg',20,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CLARO/CTE','El Salvador',706,'sv',503,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','El Salvador',706,'sv',503,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('INTELFON SA de CV','El Salvador',706,'sv',503,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telefonica','El Salvador',706,'sv',503,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telemovil','El Salvador',706,'sv',503,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HiTs-GE','Equatorial Guinea',627,'gq',240,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ORANGE/GETESA','Equatorial Guinea',627,'gq',240,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Eritel','Eritrea',657,'er',291,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('EMT GSM','Estonia',248,'ee',372,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Radiolinja Eesti','Estonia',248,'ee',372,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2 Eesti AS','Estonia',248,'ee',372,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Top Connect OU','Estonia',248,'ee',372,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ETH/MTN','Ethiopia',636,'et',251,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable and Wireless South Atlantic Ltd (Falkland Islands','Falkland Islands (Malvinas)',750,'fk',500,'001');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Edge Mobile Sp/F','Faroe Islands',288,'fo',298,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Faroese Telecom','Faroe Islands',288,'fo',298,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Kall GSM','Faroe Islands',288,'fo',298,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiCell','Fiji',542,'fj',679,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Fiji',542,'fj',679,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Alands','Finland',244,'fi',358,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Compatel Ltd','Finland',244,'fi',358,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DNA/Finnet','Finland',244,'fi',358,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DNA/Finnet','Finland',244,'fi',358,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DNA/Finnet','Finland',244,'fi',358,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DNA/Finnet','Finland',244,'fi',358,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Elisa/Saunalahti','Finland',244,'fi',358,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Elisa/Saunalahti','Finland',244,'fi',358,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ID-Mobile','Finland',244,'fi',358,'82');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mundio Mobile (Finland) Ltd','Finland',244,'fi',358,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nokia Oyj','Finland',244,'fi',358,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Oy Finland','Finland',244,'fi',358,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TeliaSonera','Finland',244,'fi',358,'91');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AFONE SA','France',208,'fr',33,'27');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Association Plate-forme Telecom','France',208,'fr',33,'92');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Astrium','France',208,'fr',33,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bouygues Telecom','France',208,'fr',33,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bouygues Telecom','France',208,'fr',33,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bouygues Telecom','France',208,'fr',33,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lliad/FREE Mobile','France',208,'fr',33,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GlobalStar','France',208,'fr',33,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GlobalStar','France',208,'fr',33,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GlobalStar','France',208,'fr',33,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','France',208,'fr',33,'29');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Legos - Local Exchange Global Operation Services SA','France',208,'fr',33,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lliad/FREE Mobile','France',208,'fr',33,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lliad/FREE Mobile','France',208,'fr',33,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile SARL','France',208,'fr',33,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MobiquiThings','France',208,'fr',33,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MobiquiThings','France',208,'fr',33,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mundio Mobile (France) Ltd','France',208,'fr',33,'31');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NRJ','France',208,'fr',33,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Virgin Mobile/Omer','France',208,'fr',33,'89');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Virgin Mobile/Omer','France',208,'fr',33,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','France',208,'fr',33,'91');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','France',208,'fr',33,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','France',208,'fr',33,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('S.F.R.','France',208,'fr',33,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('S.F.R.','France',208,'fr',33,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('S.F.R.','France',208,'fr',33,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('S.F.R.','France',208,'fr',33,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SISTEER','France',208,'fr',33,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tel/Te','France',208,'fr',33,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Transatel SA','France',208,'fr',33,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bouygues/DigiCel','French Guiana',340,'fg',594,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange Caribe','French Guiana',340,'fg',594,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Outremer Telecom','French Guiana',340,'fg',594,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TelCell GSM','French Guiana',340,'fg',594,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TelCell GSM','French Guiana',340,'fg',594,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Pacific Mobile Telecom (PMT)','French Polynesia',547,'pf',689,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vini/Tikiphone','French Polynesia',547,'pf',689,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Azur/Usan S.A.','Gabon',628,'ga',241,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Libertis S.A.','Gabon',628,'ga',241,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MOOV/Telecel','Gabon',628,'ga',241,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/ZAIN/Celtel Gabon S.A.','Gabon',628,'ga',241,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Africel','Gambia',607,'gm',220,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comium','Gambia',607,'gm',220,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Gamcel','Gambia',607,'gm',220,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Q-Cell','Gambia',607,'gm',220,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Geocell Ltd.','Georgia',282,'ge',995,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Iberiatel Ltd.','Georgia',282,'ge',995,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Magti GSM Ltd.','Georgia',282,'ge',995,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MobiTel/Beeline','Georgia',282,'ge',995,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Silknet','Georgia',282,'ge',995,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('E-Plus','Germany',262,'de',49,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DB Netz AG','Germany',262,'de',49,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Debitel','Germany',262,'de',49,'n/a');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('E-Plus','Germany',262,'de',49,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('E-Plus','Germany',262,'de',49,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('E-Plus','Germany',262,'de',49,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('E-Plus','Germany',262,'de',49,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('E-Plus','Germany',262,'de',49,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Group 3G UMTS','Germany',262,'de',49,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile','Germany',262,'de',49,'43');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobilcom','Germany',262,'de',49,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O2','Germany',262,'de',49,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O2','Germany',262,'de',49,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O2','Germany',262,'de',49,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Talkline','Germany',262,'de',49,'n/a');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-mobile/Telekom','Germany',262,'de',49,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-mobile/Telekom','Germany',262,'de',49,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telogic/ViStream','Germany',262,'de',49,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone D2','Germany',262,'de',49,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone D2','Germany',262,'de',49,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone D2','Germany',262,'de',49,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone D2','Germany',262,'de',49,'42');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Expresso Ghana Ltd','Ghana',620,'gh',233,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GloMobile','Ghana',620,'gh',233,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Milicom/Tigo','Ghana',620,'gh',233,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN','Ghana',620,'gh',233,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Ghana',620,'gh',233,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/ZAIN','Ghana',620,'gh',233,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CTS Mobile','Gibraltar',266,'gi',350,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('eazi telecom','Gibraltar',266,'gi',350,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Gibtel GSM','Gibraltar',266,'gi',350,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AMD Telecom SA','Greece',202,'gr',30,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cosmote','Greece',202,'gr',30,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cosmote','Greece',202,'gr',30,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CyTa Mobile','Greece',202,'gr',30,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Organismos Sidirodromon Ellados (OSE)','Greece',202,'gr',30,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('OTE Hellenic Telecommunications Organization SA','Greece',202,'gr',30,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tim/Wind','Greece',202,'gr',30,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tim/Wind','Greece',202,'gr',30,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Greece',202,'gr',30,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele Greenland','Greenland',290,'gl',299,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Wireless','Grenada',352,'gd',1473,'110');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Grenada',352,'gd',1473,'030');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Grenada',352,'gd',1473,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Dauphin Telecom SU (Guadeloupe Telecom)','Guadeloupe',340,'gp',590,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'Guadeloupe',340,'gp',590,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Docomo','Guam',310,'gu',1671,'470');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Docomo','Guam',310,'gu',1671,'370');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GTA Wireless','Guam',310,'gu',1671,'140');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Guam Teleph. Auth','Guam',310,'gu',1671,'033');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('IT&E OverSeas','Guam',310,'gu',1671,'032');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wave Runner LLC','Guam',311,'gu',1671,'250');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro','Guatemala',704,'gt',502,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telefonica','Guatemala',704,'gt',502,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIGO/COMCEL','Guatemala',704,'gt',502,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Areeba','Guinea',611,'gn',224,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Celcom','Guinea',611,'gn',224,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Intercel','Guinea',611,'gn',224,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Sonatel/Spacetel','Guinea',611,'gn',224,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SotelGui','Guinea',611,'gn',224,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GuineTel','Guinea-Bissau',632,'gw',245,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Guinea-Bissau',632,'gw',245,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SpaceTel','Guinea-Bissau',632,'gw',245,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cellink Plus','Guyana',738,'gy',592,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiCel','Guyana',738,'gy',592,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comcel','Haiti',372,'ht',509,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Haiti',372,'ht',509,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('National Telecom SA (NatCom)','Haiti',372,'ht',509,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Honduras',708,'hn',504,'040');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HonduTel','Honduras',708,'hn',504,'030');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SERCOM/CLARO','Honduras',708,'hn',504,'001');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telefonica/CELTEL','Honduras',708,'hn',504,'002');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Mobile/Peoples','Hongkong, China',454,'hk',852,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Mobile/Peoples','Hongkong, China',454,'hk',852,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Mobile/Peoples','Hongkong, China',454,'hk',852,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Motion','Hongkong, China',454,'hk',852,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Unicom Ltd','Hongkong, China',454,'hk',852,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China-HongKong Telecom Ltd (CHKTL)','Hongkong, China',454,'hk',852,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Citic Telecom Ltd.','Hongkong, China',454,'hk',852,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CSL Ltd.','Hongkong, China',454,'hk',852,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CSL Ltd.','Hongkong, China',454,'hk',852,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CSL Ltd.','Hongkong, China',454,'hk',852,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CSL/New World PCS Ltd.','Hongkong, China',454,'hk',852,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G/Hutchinson','Hongkong, China',454,'hk',852,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G/Hutchinson','Hongkong, China',454,'hk',852,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G/Hutchinson','Hongkong, China',454,'hk',852,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G/Hutchinson','Hongkong, China',454,'hk',852,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HKT/PCCW','Hongkong, China',454,'hk',852,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HKT/PCCW','Hongkong, China',454,'hk',852,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HKT/PCCW','Hongkong, China',454,'hk',852,'29');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HKT/PCCW','Hongkong, China',454,'hk',852,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('shared by private TETRA systems','Hongkong, China',454,'hk',852,'47');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('shared by private TETRA systems','Hongkong, China',454,'hk',852,'40');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Truephone','Hongkong, China',454,'hk',852,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone/SmarTone','Hongkong, China',454,'hk',852,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone/SmarTone','Hongkong, China',454,'hk',852,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone/SmarTone','Hongkong, China',454,'hk',852,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Pannon/Telenor','Hungary',216,'hu',36,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-mobile/Magyar','Hungary',216,'hu',36,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UPC Magyarorszag Kft.','Hungary',216,'hu',36,'71');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Hungary',216,'hu',36,'70');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Amitelo','Iceland',274,'is',354,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('IceCell','Iceland',274,'is',354,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Siminn','Iceland',274,'is',354,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Siminn','Iceland',274,'is',354,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NOVA','Iceland',274,'is',354,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VIKING/IMC','Iceland',274,'is',354,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone/Tal hf','Iceland',274,'is',354,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone/Tal hf','Iceland',274,'is',354,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone/Tal hf','Iceland',274,'is',354,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel','India',404,'in',91,'33');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel','India',404,'in',91,'29');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel','India',404,'in',91,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel','India',404,'in',91,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel','India',404,'in',91,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel','India',404,'in',91,'42');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel Digilink India','India',404,'in',91,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel Digilink India','India',404,'in',91,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircel Digilink India','India',404,'in',91,'60');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AirTel','India',405,'in',91,'53');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Barakhamba Sales & Serv.','India',404,'in',91,'86');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Barakhamba Sales & Serv.','India',404,'in',91,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'57');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'80');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'73');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'34');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'66');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'55');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'72');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'64');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'54');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'71');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'76');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'62');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'53');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'59');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'75');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'51');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'58');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'81');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'74');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BSNL','India',404,'in',91,'38');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bharti Airtel Limited (Delhi)','India',404,'in',91,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bharti Airtel Limited (Karnataka) (India)','India',404,'in',91,'045');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CellOne A&N','India',404,'in',91,'79');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Escorts Telecom Ltd.','India',404,'in',91,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Escorts Telecom Ltd.','India',404,'in',91,'87');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Escorts Telecom Ltd.','India',404,'in',91,'82');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Escorts Telecom Ltd.','India',404,'in',91,'89');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Escotel Mobile Communications','India',404,'in',91,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Escotel Mobile Communications','India',404,'in',91,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Escotel Mobile Communications','India',404,'in',91,'56');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Fascel Limited','India',405,'in',91,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Fascel','India',404,'in',91,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hexacom India','India',404,'in',91,'70');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hexcom India','India',404,'in',91,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Idea Cellular Ltd.','India',404,'in',91,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Idea Cellular Ltd.','India',404,'in',91,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Idea Cellular Ltd.','India',404,'in',91,'78');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Idea Cellular Ltd.','India',404,'in',91,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Idea Cellular Ltd.','India',404,'in',91,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mahanagar Telephone Nigam','India',404,'in',91,'69');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mahanagar Telephone Nigam','India',404,'in',91,'68');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Reliable Internet Services','India',404,'in',91,'83');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Reliance Telecom Private','India',404,'in',91,'36');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Reliance Telecom Private','India',404,'in',91,'52');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Reliance Telecom Private','India',404,'in',91,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Reliance Telecom Private','India',404,'in',91,'67');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Reliance Telecom Private','India',404,'in',91,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Reliance Telecom Private','India',404,'in',91,'85');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Reliance Telecom Private','India',404,'in',91,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('RPG Cellular','India',404,'in',91,'41');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Spice','India',404,'in',91,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Spice','India',404,'in',91,'44');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sterling Cellular Ltd.','India',404,'in',91,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TATA / Karnataka','India',405,'in',91,'034');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Usha Martin Telecom','India',404,'in',91,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Axis/Natrindo','Indonesia',510,'id',62,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Esia (PT Bakrie Telecom) (CDMA)','Indonesia',510,'id',62,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Flexi (PT Telkom) (CDMA)','Indonesia',510,'id',62,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G CP','Indonesia',510,'id',62,'89');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Indosat/Satelindo/M3','Indonesia',510,'id',62,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Indosat/Satelindo/M3','Indonesia',510,'id',62,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PT Pasifik Satelit Nusantara (PSN)','Indonesia',510,'id',62,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PT Sampoerna Telekomunikasi Indonesia (STI)','Indonesia',510,'id',62,'27');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PT Smartfren Telecom Tbk','Indonesia',510,'id',62,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PT Smartfren Telecom Tbk','Indonesia',510,'id',62,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PT. Excelcom','Indonesia',510,'id',62,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telkomsel','Indonesia',510,'id',62,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Antarctica','International Networks',901,'n/a',882,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobile Telecommunications Company of Esfahan JV-PJS (MTCE)','Iran',432,'ir',98,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTCE','Iran',432,'ir',98,'70');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/IranCell','Iran',432,'ir',98,'35');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Rightel','Iran',432,'ir',98,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Taliya','Iran',432,'ir',98,'32');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MCI/TCI','Iran',432,'ir',98,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TKC/KFZO','Iran',432,'ir',98,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Asia Cell','Iraq',418,'iq',964,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Itisaluna and Kalemat','Iraq',418,'iq',964,'92');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Korek','Iraq',418,'iq',964,'82');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Korek','Iraq',418,'iq',964,'40');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobitel (Iraq-Kurdistan) and Moutiny','Iraq',418,'iq',964,'45');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orascom Telecom','Iraq',418,'iq',964,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ZAIN/Atheer/Orascom','Iraq',418,'iq',964,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sanatel','Iraq',418,'iq',964,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Access Telecom Ltd.','Ireland',272,'ie',353,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Clever Communications Ltd','Ireland',272,'ie',353,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('eircom Ltd','Ireland',272,'ie',353,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Three/H3G','Ireland',272,'ie',353,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tesco Mobile/Liffey Telecom','Ireland',272,'ie',353,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile','Ireland',272,'ie',353,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Meteor Mobile Ltd.','Ireland',272,'ie',353,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Three/O2/Digifone','Ireland',272,'ie',353,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone Eircell','Ireland',272,'ie',353,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Alon Cellular Ltd','Israel',425,'il',972,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cellcom ltd.','Israel',425,'il',972,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Golan Telekom','Israel',425,'il',972,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Home Cellular Ltd','Israel',425,'il',972,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hot Mobile/Mirs','Israel',425,'il',972,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hot Mobile/Mirs','Israel',425,'il',972,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Partner Co. Ltd.','Israel',425,'il',972,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Pelephone','Israel',425,'il',972,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Pelephone','Israel',425,'il',972,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Rami Levy Hashikma Marketing Communications Ltd','Israel',425,'il',972,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telzar/AZI','Israel',425,'il',972,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BT Italia SpA','Italy',222,'it',39,'34');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Elsacom','Italy',222,'it',39,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Fastweb SpA','Italy',222,'it',39,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Fix Line','Italy',222,'it',39,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hi3G','Italy',222,'it',39,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('IPSE 2000','Italy',222,'it',39,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile Srl','Italy',222,'it',39,'35');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Noverca Italia Srl','Italy',222,'it',39,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PosteMobile SpA','Italy',222,'it',39,'33');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Premium Number(s)','Italy',222,'it',39,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('RFI Rete Ferroviaria Italiana SpA','Italy',222,'it',39,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecom Italia Mobile SpA','Italy',222,'it',39,'48');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecom Italia Mobile SpA','Italy',222,'it',39,'43');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIM','Italy',222,'it',39,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Italy',222,'it',39,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Italy',222,'it',39,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VOIP Line','Italy',222,'it',39,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('WIND (Blu) -','Italy',222,'it',39,'44');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('WIND (Blu) -','Italy',222,'it',39,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aircomm SA','Ivory Coast',612,'ci',225,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Atlantik Tel./Moov','Ivory Coast',612,'ci',225,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comium','Ivory Coast',612,'ci',225,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comstar','Ivory Coast',612,'ci',225,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN','Ivory Coast',612,'ci',225,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Ivory Coast',612,'ci',225,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('OriCell','Ivory Coast',612,'ci',225,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Wireless','Jamaica',338,'jm',1876,'110');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Wireless','Jamaica',338,'jm',1876,'020');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Wireless','Jamaica',338,'jm',1876,'180');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DIGICEL/Mossel','Jamaica',338,'jm',1876,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Y-Mobile','Japan',440,'jp',81,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'74');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'70');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'51');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'89');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'75');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'56');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'52');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',441,'jp',81,'70');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'76');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'71');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'53');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'72');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'54');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'79');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'73');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'55');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'45');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'68');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'98');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'42');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'63');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'38');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'44');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'69');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'34');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'64');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'37');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'43');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'27');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'87');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'31');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'65');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'36');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'92');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'58');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'61');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'91');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'32');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'66');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'35');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'93');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'40');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'49');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'29');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'90');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'33');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'60');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'67');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'94');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',441,'jp',81,'41');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'62');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'39');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTT Docomo','Japan',440,'jp',81,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Okinawa Cellular Telephone','Japan',440,'jp',81,'78');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'45');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'96');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'40');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',441,'jp',81,'63');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'47');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'95');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'41');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',441,'jp',81,'64');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'46');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'97');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'42');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',441,'jp',81,'65');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'90');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'92');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'98');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'43');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'93');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'48');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',441,'jp',81,'61');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'44');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'94');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',440,'jp',81,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SoftBank Mobile Corp','Japan',441,'jp',81,'62');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'85');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'83');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'86');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'81');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'80');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'84');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KDDI Corporation','Japan',440,'jp',81,'82');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Petra','Jordan',416,'jo',962,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Umniah Mobile Co.','Jordan',416,'jo',962,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Xpress','Jordan',416,'jo',962,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ZAIN /J.M.T.S','Jordan',416,'jo',962,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Beeline/KaR-Tel LLP','Kazakhstan',401,'kz',7,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Dalacom/Altel','Kazakhstan',401,'kz',7,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('K-Cell','Kazakhstan',401,'kz',7,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2/NEO/MTS','Kazakhstan',401,'kz',7,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Econet Wireless','Kenya',639,'ke',254,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Kenya',639,'ke',254,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Safaricom Ltd.','Kenya',639,'ke',254,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Zain/Celtel Ltd.','Kenya',639,'ke',254,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Kiribati Frigate','Kiribati',545,'ki',686,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sun Net','Korea N., Dem. People''s Rep.',467,'kp',850,'193');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KT Freetel Co. Ltd.','Korea S, Republic of',450,'kr',82,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KT Freetel Co. Ltd.','Korea S, Republic of',450,'kr',82,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KT Freetel Co. Ltd.','Korea S, Republic of',450,'kr',82,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('LG Telecom','Korea S, Republic of',450,'kr',82,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SK Telecom','Korea S, Republic of',450,'kr',82,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SK Telecom Co. Ltd','Korea S, Republic of',450,'kr',82,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Viva','Kuwait',419,'kw',965,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wataniya','Kuwait',419,'kw',965,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Zain','Kuwait',419,'kw',965,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AkTel LLC','Kyrgyzstan',437,'kg',996,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Beeline/Bitel','Kyrgyzstan',437,'kg',996,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MEGACOM','Kyrgyzstan',437,'kg',996,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O!/NUR Telecom','Kyrgyzstan',437,'kg',996,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ETL Mobile','Laos P.D.R.',457,'la',856,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lao Tel','Laos P.D.R.',457,'la',856,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Beeline/Tigo/Millicom','Laos P.D.R.',457,'la',856,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UNITEL/LAT','Laos P.D.R.',457,'la',856,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bite','Latvia',247,'lv',371,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Latvian Mobile Phone','Latvia',247,'lv',371,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SIA Camel Mobile','Latvia',247,'lv',371,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SIA IZZI','Latvia',247,'lv',371,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SIA Master Telecom','Latvia',247,'lv',371,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SIA Rigatta','Latvia',247,'lv',371,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2','Latvia',247,'lv',371,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TRIATEL/Telekom Baltija','Latvia',247,'lv',371,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cellis','Lebanon',415,'lb',961,'35');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cellis','Lebanon',415,'lb',961,'33');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cellis','Lebanon',415,'lb',961,'32');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('FTML Cellis','Lebanon',415,'lb',961,'34');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MIC2/LibanCell/MTC','Lebanon',415,'lb',961,'39');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MIC2/LibanCell/MTC','Lebanon',415,'lb',961,'38');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MIC2/LibanCell/MTC','Lebanon',415,'lb',961,'37');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MIC1 (Alfa)','Lebanon',415,'lb',961,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MIC2/LibanCell/MTC','Lebanon',415,'lb',961,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MIC2/LibanCell/MTC','Lebanon',415,'lb',961,'36');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Econet/Ezi-cel','Lesotho',651,'ls',266,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodacom Lesotho','Lesotho',651,'ls',266,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CELLCOM','Liberia',618,'lr',231,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comium BVI','Liberia',618,'lr',231,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Libercell','Liberia',618,'lr',231,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('LibTelco','Liberia',618,'lr',231,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lonestar','Liberia',618,'lr',231,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Al-Madar','Libya',606,'ly',218,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Al-Madar','Libya',606,'ly',218,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hatef','Libya',606,'ly',218,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Libyana','Libya',606,'ly',218,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Libyana','Libya',606,'ly',218,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CUBIC (Liechtenstein','Liechtenstein',295,'li',423,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('First Mobile AG','Liechtenstein',295,'li',423,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Liechtenstein',295,'li',423,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Swisscom FL AG','Liechtenstein',295,'li',423,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Alpmobile/Tele2','Liechtenstein',295,'li',423,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecom FL1 AG','Liechtenstein',295,'li',423,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bite','Lithuania',246,'lt',370,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Omnitel','Lithuania',246,'lt',370,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2','Lithuania',246,'lt',370,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Millicom Tango GSM','Luxembourg',270,'lu',352,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('P+T/Post LUXGSM','Luxembourg',270,'lu',352,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/VOXmobile S.A.','Luxembourg',270,'lu',352,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('C.T.M. TELEMOVEL+','Macao, China',455,'mo',853,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('C.T.M. TELEMOVEL+','Macao, China',455,'mo',853,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('China Telecom','Macao, China',455,'mo',853,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hutchison Telephone Co. Ltd','Macao, China',455,'mo',853,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hutchison Telephone Co. Ltd','Macao, China',455,'mo',853,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smartone Mobile','Macao, China',455,'mo',853,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smartone Mobile','Macao, China',455,'mo',853,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ONE/Cosmofone','Macedonia',294,'mk',389,'75');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ONE/Cosmofone','Macedonia',294,'mk',389,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile/Mobimak','Macedonia',294,'mk',389,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VIP Mobile','Macedonia',294,'mk',389,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/MADACOM','Madagascar',646,'mg',261,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Soci','Madagascar',646,'mg',261,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sacel','Madagascar',646,'mg',261,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telma','Madagascar',646,'mg',261,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TNM/Telekom Network Ltd.','Malawi',650,'mw',265,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Zain/Celtel ltd.','Malawi',650,'mw',265,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Art900','Malaysia',502,'my',60,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Baraka Telecom Sdn Bhd','Malaysia',502,'my',60,'151');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CelCom','Malaysia',502,'my',60,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CelCom','Malaysia',502,'my',60,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CelCom','Malaysia',502,'my',60,'198');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digi Telecommunications','Malaysia',502,'my',60,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digi Telecommunications','Malaysia',502,'my',60,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Electcoms Wireless Sdn Bhd','Malaysia',502,'my',60,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Maxis','Malaysia',502,'my',60,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Maxis','Malaysia',502,'my',60,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTX Utara','Malaysia',502,'my',60,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Webe/Packet One Networks (Malaysia) Sdn Bhd','Malaysia',502,'my',60,'153');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Samata Communications Sdn Bhd','Malaysia',502,'my',60,'155');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tron/Talk Focus Sdn Bhd','Malaysia',502,'my',60,'154');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('U Mobile','Malaysia',502,'my',60,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('XOX Com Sdn Bhd','Malaysia',502,'my',60,'195');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('YES','Malaysia',502,'my',60,'152');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Dhiraagu/C&W','Maldives',472,'mv',960,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ooredo/Wataniya','Maldives',472,'mv',960,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Malitel','Mali',610,'ml',223,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/IKATEL','Mali',610,'ml',223,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GO Mobile','Malta',278,'mt',356,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Melita','Malta',278,'mt',356,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Malta',278,'mt',356,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UTS Caraibe','Martinique (French Department of)',340,'mq',596,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Chinguitel SA','Mauritania',609,'mr',222,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mattel','Mauritania',609,'mr',222,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mauritel','Mauritania',609,'mr',222,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Emtel Ltd','Mauritius',617,'mu',230,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mahanagar Telephone','Mauritius',617,'mu',230,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mahanagar Telephone','Mauritius',617,'mu',230,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Cellplus','Mauritius',617,'mu',230,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T/IUSACell','Mexico',334,'mx',52,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T/IUSACell','Mexico',334,'mx',52,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T/IUSACell','Mexico',334,'mx',52,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T/IUSACell','Mexico',334,'mx',52,'040');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar/Pegaso','Mexico',334,'mx',52,'030');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar/Pegaso','Mexico',334,'mx',52,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NEXTEL','Mexico',334,'mx',52,'010');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NEXTEL','Mexico',334,'mx',52,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NEXTEL','Mexico',334,'mx',52,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NEXTEL','Mexico',334,'mx',52,'090');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Operadora Unefon SA de CV','Mexico',334,'mx',52,'080');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Operadora Unefon SA de CV','Mexico',334,'mx',52,'070');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SAI PCS','Mexico',334,'mx',52,'060');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TelCel/America Movil','Mexico',334,'mx',52,'020');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TelCel/America Movil','Mexico',334,'mx',52,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('FSM Telecom','Micronesia',550,'fm',691,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Eventis Mobile','Moldova',259,'md',373,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('IDC/Unite','Moldova',259,'md',373,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('IDC/Unite','Moldova',259,'md',373,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('IDC/Unite','Moldova',259,'md',373,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Moldcell','Moldova',259,'md',373,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Voxtel','Moldova',259,'md',373,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Monaco Telecom','Monaco',212,'mc',377,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Monaco Telecom','Monaco',212,'mc',377,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('G-Mobile Corporation Ltd','Mongolia',428,'mn',976,'98');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobicom','Mongolia',428,'mn',976,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Skytel Co. Ltd','Mongolia',428,'mn',976,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Skytel Co. Ltd','Mongolia',428,'mn',976,'91');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unitel','Mongolia',428,'mn',976,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Monet/T-mobile','Montenegro',297,'me',382,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mtel','Montenegro',297,'me',382,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor/Promonte GSM','Montenegro',297,'me',382,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Wireless','Montserrat',354,'ms',1664,'860');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('IAM/Itissallat','Morocco',604,'ma',212,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('INWI/WANA','Morocco',604,'ma',212,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Medi Telecom','Morocco',604,'ma',212,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('mCel','Mozambique',643,'mz',258,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movitel','Mozambique',643,'mz',258,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodacom','Mozambique',643,'mz',258,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Myanmar Post & Teleco.','Myanmar (Burma)',414,'mm',95,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Oreedoo','Myanmar (Burma)',414,'mm',95,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor','Myanmar (Burma)',414,'mm',95,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Leo / Orascom','Namibia',649,'na',264,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTC','Namibia',649,'na',264,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Switch/Nam. Telec.','Namibia',649,'na',264,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ncell','Nepal',429,'np',977,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NT Mobile / Namaste','Nepal',429,'np',977,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smart Cell','Nepal',429,'np',977,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('6GMOBILE BV','Netherlands',204,'nl',31,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aspider Solutions','Netherlands',204,'nl',31,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Elephant Talk Communications Premium Rate Services Netherlands BV','Netherlands',204,'nl',31,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Intercity Mobile Communications BV','Netherlands',204,'nl',31,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KPN Telecom B.V.','Netherlands',204,'nl',31,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KPN Telecom B.V.','Netherlands',204,'nl',31,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KPN Telecom B.V.','Netherlands',204,'nl',31,'69');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KPN/Telfort','Netherlands',204,'nl',31,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lancelot BV','Netherlands',204,'nl',31,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile Ltd','Netherlands',204,'nl',31,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mundio/Vectone Mobile','Netherlands',204,'nl',31,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NS Railinfrabeheer B.V.','Netherlands',204,'nl',31,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Private Mobility Nederland BV','Netherlands',204,'nl',31,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile B.V.','Netherlands',204,'nl',31,'98');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile B.V.','Netherlands',204,'nl',31,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-mobile/former Orange','Netherlands',204,'nl',31,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2','Netherlands',204,'nl',31,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Teleena Holding BV','Netherlands',204,'nl',31,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unify Mobile','Netherlands',204,'nl',31,'68');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UPC Nederland BV','Netherlands',204,'nl',31,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone Libertel','Netherlands',204,'nl',31,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Voiceworks Mobile BV','Netherlands',204,'nl',31,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ziggo BV','Netherlands',204,'nl',31,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cingular Wireless','Netherlands Antilles',362,'an',599,'630');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TELCELL GSM','Netherlands Antilles',362,'an',599,'51');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SETEL GSM','Netherlands Antilles',362,'an',599,'91');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UTS Wireless','Netherlands Antilles',362,'an',599,'951');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('OPT Mobilis','New Caledonia',546,'nc',687,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('2degrees','New Zealand',530,'nz',64,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Spark/NZ Telecom','New Zealand',530,'nz',64,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Spark/NZ Telecom','New Zealand',530,'nz',64,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telstra','New Zealand',530,'nz',64,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Two Degrees Mobile Ltd','New Zealand',530,'nz',64,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','New Zealand',530,'nz',64,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Walker Wireless Ltd.','New Zealand',530,'nz',64,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Empresa Nicaraguense de Telecomunicaciones SA (ENITEL)','Nicaragua',710,'ni',505,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Nicaragua',710,'ni',505,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro','Nicaragua',710,'ni',505,'73');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MOOV/TeleCel','Niger',614,'ne',227,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Sahelc.','Niger',614,'ne',227,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Sahelc.','Niger',614,'ne',227,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Zain/CelTel','Niger',614,'ne',227,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/ZAIN/Econet','Nigeria',621,'ng',234,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ETISALAT','Nigeria',621,'ng',234,'60');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Glo Mobile','Nigeria',621,'ng',234,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('M-Tel/Nigeria Telecom. Ltd.','Nigeria',621,'ng',234,'40');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN','Nigeria',621,'ng',234,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Starcomms','Nigeria',621,'ng',234,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Visafone','Nigeria',621,'ng',234,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Visafone','Nigeria',621,'ng',234,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Niue Telecom','Niue',555,'nu',683,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Com4 AS','Norway',242,'no',47,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ICE Nordisk Mobiltelefon AS','Norway',242,'no',47,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Jernbaneverket (GSM-R)','Norway',242,'no',47,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Jernbaneverket (GSM-R)','Norway',242,'no',47,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile Ltd','Norway',242,'no',47,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Netcom','Norway',242,'no',47,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Network Norway AS','Norway',242,'no',47,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Network Norway AS','Norway',242,'no',47,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ICE Nordisk Mobiltelefon AS','Norway',242,'no',47,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Mobil A/S','Norway',242,'no',47,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2','Norway',242,'no',47,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor','Norway',242,'no',47,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor','Norway',242,'no',47,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Teletopia','Norway',242,'no',47,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ventelo AS','Norway',242,'no',47,'017');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ventelo AS','Norway',242,'no',47,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nawras','Oman',422,'om',968,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Oman Mobile/GTO','Oman',422,'om',968,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Instaphone','Pakistan',410,'pk',92,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobilink','Pakistan',410,'pk',92,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor','Pakistan',410,'pk',92,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UFONE/PAKTel','Pakistan',410,'pk',92,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Warid Telecom','Pakistan',410,'pk',92,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ZONG/CMPak','Pakistan',410,'pk',92,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Palau Mobile Corp. (PMC) (Palau','Palau (Republic of)',552,'pw',680,'80');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Palau National Communications Corp. (PNCC) (Palau','Palau (Republic of)',552,'pw',680,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Jawwal','Palestinian Territory',425,'ps',970,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wataniya Mobile','Palestinian Territory',425,'ps',970,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & W./Mas Movil','Panama',714,'pa',507,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro','Panama',714,'pa',507,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Panama',714,'pa',507,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Panama',714,'pa',507,'020');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Panama',714,'pa',507,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Papua New Guinea',537,'pg',675,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GreenCom PNG Ltd','Papua New Guinea',537,'pg',675,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Pacific Mobile','Papua New Guinea',537,'pg',675,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/Hutchison','Paraguay',744,'py',595,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Compa','Paraguay',744,'py',595,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hola/VOX','Paraguay',744,'py',595,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIM/Nucleo/Personal','Paraguay',744,'py',595,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tigo/Telecel','Paraguay',744,'py',595,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro /Amer.Mov./TIM','Peru',716,'pe',51,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro /Amer.Mov./TIM','Peru',716,'pe',51,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GlobalStar','Peru',716,'pe',51,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GlobalStar','Peru',716,'pe',51,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Peru',716,'pe',51,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel','Peru',716,'pe',51,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nextel','Peru',716,'pe',51,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Viettel Mobile','Peru',716,'pe',51,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Fix Line','Philippines',515,'ph',63,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Globe Telecom','Philippines',515,'ph',63,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Globe Telecom','Philippines',515,'ph',63,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Next Mobile','Philippines',515,'ph',63,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('RED Mobile/Cure','Philippines',515,'ph',63,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smart','Philippines',515,'ph',63,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SUN/Digitel','Philippines',515,'ph',63,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aero2 SP','Poland',260,'pl',48,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AMD Telecom.','Poland',260,'pl',48,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CallFreedom Sp. z o.o.','Poland',260,'pl',48,'38');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cyfrowy POLSAT S.A.','Poland',260,'pl',48,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('e-Telko','Poland',260,'pl',48,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile','Poland',260,'pl',48,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobyland','Poland',260,'pl',48,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mundio Mobile Sp. z o.o.','Poland',260,'pl',48,'36');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Play/P4','Poland',260,'pl',48,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NORDISK Polska','Poland',260,'pl',48,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/IDEA/Centertel','Poland',260,'pl',48,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/IDEA/Centertel','Poland',260,'pl',48,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PKP Polskie Linie Kolejowe S.A.','Poland',260,'pl',48,'35');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Play/P4','Poland',260,'pl',48,'98');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Play/P4','Poland',260,'pl',48,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Polkomtel/Plus','Poland',260,'pl',48,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sferia','Poland',260,'pl',48,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sferia','Poland',260,'pl',48,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sferia','Poland',260,'pl',48,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile/ERA','Poland',260,'pl',48,'34');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile/ERA','Poland',260,'pl',48,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2','Poland',260,'pl',48,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2','Poland',260,'pl',48,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile','Portugal',268,'pt',351,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NOS/Optimus','Portugal',268,'pt',351,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NOS/Optimus','Portugal',268,'pt',351,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MEO/TMN','Portugal',268,'pt',351,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Portugal',268,'pt',351,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Puerto Rico Telephone Company Inc. (PRTC)','Puerto Rico',330,'pr',NULL,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Puerto Rico Telephone Company Inc. (PRTC)','Puerto Rico',330,'pr',NULL,'110');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ooredoo/Qtel','Qatar',427,'qa',974,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Qatar',427,'qa',974,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Reunion',647,'re',262,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Outremer Telecom','Reunion',647,'re',262,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SFR','Reunion',647,'re',262,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cosmote','Romania',226,'ro',40,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Enigma Systems','Romania',226,'ro',40,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile','Romania',226,'ro',40,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Romania',226,'ro',40,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('RCS&RDS Digi Mobile','Romania',226,'ro',40,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Romtelecom SA','Romania',226,'ro',40,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telemobil/Zapp','Romania',226,'ro',40,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Romania',226,'ro',40,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telemobil/Zapp','Romania',226,'ro',40,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Baykal Westcom','Russian Federation',250,'ru',79,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BeeLine/VimpelCom','Russian Federation',250,'ru',79,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DTC/Don Telecom','Russian Federation',250,'ru',79,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Kuban GSM','Russian Federation',250,'ru',79,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MOTIV/LLC Ekaterinburg-2000','Russian Federation',250,'ru',79,'35');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Megafon','Russian Federation',250,'ru',79,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTS','Russian Federation',250,'ru',79,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NCC','Russian Federation',250,'ru',79,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NTC','Russian Federation',250,'ru',79,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('OJSC Altaysvyaz','Russian Federation',250,'ru',79,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orensot','Russian Federation',250,'ru',79,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Printelefone','Russian Federation',250,'ru',79,'92');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sibchallenge','Russian Federation',250,'ru',79,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('StavTelesot','Russian Federation',250,'ru',79,'44');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2/ECC/Volgogr.','Russian Federation',250,'ru',79,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecom XXL','Russian Federation',250,'ru',79,'93');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UralTel','Russian Federation',250,'ru',79,'39');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UralTel','Russian Federation',250,'ru',79,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BeeLine/VimpelCom','Russian Federation',250,'ru',79,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Yenisey Telecom','Russian Federation',250,'ru',79,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ZAO SMARTS','Russian Federation',250,'ru',79,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ZAO SMARTS','Russian Federation',250,'ru',79,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel','Rwanda',635,'rw',250,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Rwandacell','Rwanda',635,'rw',250,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIGO','Rwanda',635,'rw',250,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Wireless','Saint Kitts and Nevis',356,'kn',1869,'110');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Saint Kitts and Nevis',356,'kn',1869,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UTS Cariglobe','Saint Kitts and Nevis',356,'kn',1869,'70');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Wireless','Saint Lucia',358,'lc',1758,'110');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cingular Wireless','Saint Lucia',358,'lc',1758,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel (St Lucia) Limited','Saint Lucia',358,'lc',1758,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Samoatel Mobile','Samoa',549,'ws',685,'27');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecom Samoa Cellular Ltd.','Samoa',549,'ws',685,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Prima Telecom','San Marino',292,'sm',378,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CSTmovel','Sao Tome & Principe',626,'st',239,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AeroMobile','Satellite Networks',901,'n/a',870,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('InMarSAT','Satellite Networks',901,'n/a',870,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Maritime Communications Partner AS','Satellite Networks',901,'n/a',870,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Thuraya Satellite','Satellite Networks',901,'n/a',870,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Zain','Saudi Arabia',420,'sa',966,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Etihad/Etisalat/Mobily','Saudi Arabia',420,'sa',966,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lebara Mobile','Saudi Arabia',420,'sa',966,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('STC/Al Jawal','Saudi Arabia',420,'sa',966,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Virgin Mobile','Saudi Arabia',420,'sa',966,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Zain','Saudi Arabia',420,'sa',966,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Expresso/Sudatel','Senegal',608,'sn',221,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange/Sonatel','Senegal',608,'sn',221,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIGO/Sentel GSM','Senegal',608,'sn',221,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTS/Telekom Srbija','Serbia',220,'rs',381,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor/Mobtel','Serbia',220,'rs',381,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor/Mobtel','Serbia',220,'rs',381,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VIP Mobile','Serbia',220,'rs',381,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel','Seychelles',633,'sc',248,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('C&W','Seychelles',633,'sc',248,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smartcom','Seychelles',633,'sc',248,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Africel','Sierra Leone',619,'sl',232,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Zain/Celtel','Sierra Leone',619,'sl',232,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comium','Sierra Leone',619,'sl',232,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Africel','Sierra Leone',619,'sl',232,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tigo/Millicom','Sierra Leone',619,'sl',232,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobitel','Sierra Leone',619,'sl',232,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GRID Communications Pte Ltd','Singapore',525,'sg',65,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MobileOne Ltd','Singapore',525,'sg',65,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Singtel','Singapore',525,'sg',65,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Singtel','Singapore',525,'sg',65,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Singtel','Singapore',525,'sg',65,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Starhub','Singapore',525,'sg',65,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Starhub','Singapore',525,'sg',65,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('4Ka','Slovakia',231,'sk',421,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O2','Slovakia',231,'sk',421,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Slovakia',231,'sk',421,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Slovakia',231,'sk',421,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Slovakia',231,'sk',421,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','Slovakia',231,'sk',421,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','Slovakia',231,'sk',421,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Zeleznice Slovenskej republiky (ZSR)','Slovakia',231,'sk',421,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobitel','Slovenia',293,'si',386,'41');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SI.Mobil','Slovenia',293,'si',386,'40');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Slovenske zeleznice d.o.o.','Slovenia',293,'si',386,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-2 d.o.o.','Slovenia',293,'si',386,'64');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telemach/TusMobil/VEGA','Slovenia',293,'si',386,'70');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('bemobile','Solomon Islands',540,'sb',677,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BREEZE','Solomon Islands',540,'sb',677,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BREEZE','Solomon Islands',540,'sb',677,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Golis','Somalia',637,'so',252,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HorTel','Somalia',637,'so',252,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nationlink','Somalia',637,'so',252,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nationlink','Somalia',637,'so',252,'60');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Somafone','Somalia',637,'so',252,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Somtel','Somalia',637,'so',252,'82');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Somtel','Somalia',637,'so',252,'71');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telesom','Somalia',637,'so',252,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('8.ta','South Africa',655,'za',27,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cape Town Metropolitan','South Africa',655,'za',27,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cell C','South Africa',655,'za',27,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN','South Africa',655,'za',27,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN','South Africa',655,'za',27,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sentech','South Africa',655,'za',27,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodacom','South Africa',655,'za',27,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wireless Business Solutions (Pty) Ltd','South Africa',655,'za',27,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Gemtel Ltd (South Sudan','South Sudan (Republic of)',659,'ss',NULL,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN South Sudan (South Sudan','South Sudan (Republic of)',659,'ss',NULL,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Network of The World Ltd (NOW) (South Sudan','South Sudan (Republic of)',659,'ss',NULL,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Zain South Sudan (South Sudan','South Sudan (Republic of)',659,'ss',NULL,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile SL','Spain',214,'es',34,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digi Spain Telecom SL','Spain',214,'es',34,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BT Espana  SAU','Spain',214,'es',34,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cableuropa SAU (ONO)','Spain',214,'es',34,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Euskaltel SA','Spain',214,'es',34,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('fonYou Wireless SL','Spain',214,'es',34,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ION Mobile','Spain',214,'es',34,'32');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Jazz Telecom SAU','Spain',214,'es',34,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lleida','Spain',214,'es',34,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile SL','Spain',214,'es',34,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Spain',214,'es',34,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar','Spain',214,'es',34,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Spain',214,'es',34,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Spain',214,'es',34,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Spain',214,'es',34,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('R Cable y Telec. Galicia SA','Spain',214,'es',34,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Simyo/KPN','Spain',214,'es',34,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecable de Asturias SA','Spain',214,'es',34,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Truphone','Spain',214,'es',34,'27');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','Spain',214,'es',34,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone Enabler Espana SL','Spain',214,'es',34,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Yoigo','Spain',214,'es',34,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel','Sri Lanka',413,'lk',94,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Etisalat/Tigo','Sri Lanka',413,'lk',94,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G Hutchison','Sri Lanka',413,'lk',94,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobitel Ltd.','Sri Lanka',413,'lk',94,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Dialog','Sri Lanka',413,'lk',94,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ameris','St. Pierre & Miquelon',308,'pm',508,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('C & W','St. Vincent & Gren.',360,'vc',1784,'110');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cingular','St. Vincent & Gren.',360,'vc',1784,'100');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cingular','St. Vincent & Gren.',360,'vc',1784,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','St. Vincent & Gren.',360,'vc',1784,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','St. Vincent & Gren.',360,'vc',1784,'70');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Canar Telecom','Sudan',634,'sd',249,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN','Sudan',634,'sd',249,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN','Sudan',634,'sd',249,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sudani One','Sudan',634,'sd',249,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sudani One','Sudan',634,'sd',249,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivacell','Sudan',634,'sd',249,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vivacell','Sudan',634,'sd',249,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ZAIN/Mobitel','Sudan',634,'sd',249,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ZAIN/Mobitel','Sudan',634,'sd',249,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Suriname',746,'sr',597,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telesur','Suriname',746,'sr',597,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecommunicatiebedrijf Suriname (TELESUR)','Suriname',746,'sr',597,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UNIQA','Suriname',746,'sr',597,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Swazi MTN','Swaziland',653,'sz',268,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SwaziTelecom','Swaziland',653,'sz',268,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('42 Telecom AB','Sweden',240,'se',46,'35');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('42 Telecom AB','Sweden',240,'se',46,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Beepsend','Sweden',240,'se',46,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NextGen Mobile Ltd (CardBoardFish)','Sweden',240,'se',46,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CoolTEL Aps','Sweden',240,'se',46,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digitel Mobile Srl','Sweden',240,'se',46,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Eu Tel AB','Sweden',240,'se',46,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Fogg Mobile AB','Sweden',240,'se',46,'27');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Generic Mobile Systems Sweden AB','Sweden',240,'se',46,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Gotalandsnatet AB','Sweden',240,'se',46,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G Access AB','Sweden',240,'se',46,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G Access AB','Sweden',240,'se',46,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ID Mobile','Sweden',240,'se',46,'36');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Infobip Ltd.','Sweden',240,'se',46,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lindholmen Science Park AB','Sweden',240,'se',46,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile Ltd','Sweden',240,'se',46,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mercury International Carrier Services','Sweden',240,'se',46,'29');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mundio Mobile (Sweden) Ltd','Sweden',240,'se',46,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Spring Mobil AB','Sweden',240,'se',46,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Svenska UMTS-N','Sweden',240,'se',46,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Sverige AB','Sweden',240,'se',46,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tele2 Sverige AB','Sweden',240,'se',46,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor (Vodafone)','Sweden',240,'se',46,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor (Vodafone)','Sweden',240,'se',46,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telenor (Vodafone)','Sweden',240,'se',46,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telia Mobile','Sweden',240,'se',46,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ventelo Sverige AB','Sweden',240,'se',46,'13');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wireless Maingate AB','Sweden',240,'se',46,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wireless Maingate Nordic AB','Sweden',240,'se',46,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BebbiCell AG','Switzerland',228,'ch',41,'51');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comfone AG','Switzerland',228,'ch',41,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Comfone AG','Switzerland',228,'ch',41,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Sunrise','Switzerland',228,'ch',41,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile AG','Switzerland',228,'ch',41,'54');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mundio Mobile AG','Switzerland',228,'ch',41,'52');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Salt/Orange','Switzerland',228,'ch',41,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Swisscom','Switzerland',228,'ch',41,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Sunrise','Switzerland',228,'ch',41,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Sunrise','Switzerland',228,'ch',41,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TDC Sunrise','Switzerland',228,'ch',41,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('upc cablecom GmbH','Switzerland',228,'ch',41,'53');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Spacetel','Syrian Arab Republic',417,'sy',963,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Syriatel Holdings','Syrian Arab Republic',417,'sy',963,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Syriatel Holdings','Syrian Arab Republic',417,'sy',963,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ACeS Taiwan - ACeS Taiwan Telecommunications Co Ltd','Taiwan',466,'tw',886,'68');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Asia Pacific Telecom Co. Ltd (APT)','Taiwan',466,'tw',886,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Chunghwa Telecom LDM','Taiwan',466,'tw',886,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Chunghwa Telecom LDM','Taiwan',466,'tw',886,'92');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Far EasTone','Taiwan',466,'tw',886,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Far EasTone','Taiwan',466,'tw',886,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Far EasTone','Taiwan',466,'tw',886,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Far EasTone','Taiwan',466,'tw',886,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Far EasTone','Taiwan',466,'tw',886,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Global Mobile Corp.','Taiwan',466,'tw',886,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('International Telecom Co. Ltd (FITEL)','Taiwan',466,'tw',886,'56');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KG Telecom','Taiwan',466,'tw',886,'88');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TransAsia','Taiwan',466,'tw',886,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Taiwan Cellular','Taiwan',466,'tw',886,'97');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobitai','Taiwan',466,'tw',886,'93');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Star/VIBO','Taiwan',466,'tw',886,'89');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VMAX Telecom Co. Ltd','Taiwan',466,'tw',886,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Babilon-M','Tajikistan',436,'tk',992,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bee Line','Tajikistan',436,'tk',992,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('CJSC Indigo Tajikistan','Tajikistan',436,'tk',992,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tcell/JC Somoncom','Tajikistan',436,'tk',992,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MLT/TT mobile','Tajikistan',436,'tk',992,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tcell/JC Somoncom','Tajikistan',436,'tk',992,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Benson Informatics Ltd','Tanzania',640,'tz',255,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Dovetel (T) Ltd','Tanzania',640,'tz',255,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Halotel/Viettel Ltd','Tanzania',640,'tz',255,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smile Communications Tanzania Ltd','Tanzania',640,'tz',255,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tanzania Telecommunications Company Ltd (TTCL)','Tanzania',640,'tz',255,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TIGO/MIC','Tanzania',640,'tz',255,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tri Telecomm. Ltd.','Tanzania',640,'tz',255,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodacom Ltd','Tanzania',640,'tz',255,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/ZAIN/Celtel','Tanzania',640,'tz',255,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Zantel/Zanzibar Telecom','Tanzania',640,'tz',255,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ACeS Thailand - ACeS Regional Services Co Ltd','Thailand',520,'th',66,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('ACT Mobile','Thailand',520,'th',66,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Advanced Wireless Networks/AWN','Thailand',520,'th',66,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AIS/Advanced Info Service','Thailand',520,'th',66,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digital Phone Co.','Thailand',520,'th',66,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Hutch/CAT CDMA','Thailand',520,'th',66,'00');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Total Access (DTAC)','Thailand',520,'th',66,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Total Access (DTAC)','Thailand',520,'th',66,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('True Move/Orange','Thailand',520,'th',66,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('True Move/Orange','Thailand',520,'th',66,'99');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telin/ Telkomcel','Timor-Leste',514,'tp',670,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Timor Telecom','Timor-Leste',514,'tp',670,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecel/MOOV','Togo',615,'tg',228,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecel/MOOV','Togo',615,'tg',228,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Togo Telecom/TogoCELL','Togo',615,'tg',228,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Shoreline Communication','Tonga',539,'to',676,'43');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tonga Communications','Tonga',539,'to',676,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bmobile/TSTT','Trinidad and Tobago',374,'tt',1868,'120');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bmobile/TSTT','Trinidad and Tobago',374,'tt',1868,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Trinidad and Tobago',374,'tt',1868,'130');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('LaqTel Ltd.','Trinidad and Tobago',374,'tt',1868,'140');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Tunisia',605,'tn',216,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Oreedo/Orascom','Tunisia',605,'tn',216,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TuniCell/Tunisia Telecom','Tunisia',605,'tn',216,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TuniCell/Tunisia Telecom','Tunisia',605,'tn',216,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AVEA/Aria','Turkey',286,'tr',90,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AVEA/Aria','Turkey',286,'tr',90,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Turkcell','Turkey',286,'tr',90,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone-Telsim','Turkey',286,'tr',90,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTS/Barash Communication','Turkmenistan',438,'tm',993,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Altyn Asyr/TM-Cell','Turkmenistan',438,'tm',993,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Wireless (TCI) Ltd','Turks and Caicos Islands',376,'tc',NULL,'350');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel TCI Ltd','Turks and Caicos Islands',376,'tc',NULL,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('IslandCom Communications Ltd.','Turks and Caicos Islands',376,'tc',NULL,'352');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tuvalu Telecommunication Corporation (TTC)','Tuvalu',553,'tv',NULL,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Celtel','Uganda',641,'ug',256,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('i-Tel Ltd','Uganda',641,'ug',256,'66');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('K2 Telecom Ltd','Uganda',641,'ug',256,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN Ltd.','Uganda',641,'ug',256,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Orange','Uganda',641,'ug',256,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smile Communications Uganda Ltd','Uganda',641,'ug',256,'33');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Suretelecom Uganda Ltd','Uganda',641,'ug',256,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Uganda Telecom Ltd.','Uganda',641,'ug',256,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Warid','Uganda',641,'ug',256,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Astelit/LIFE','Ukraine',255,'ua',380,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Golden Telecom','Ukraine',255,'ua',380,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Golden Telecom','Ukraine',255,'ua',380,'39');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Intertelecom Ltd (IT)','Ukraine',255,'ua',380,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KyivStar','Ukraine',255,'ua',380,'67');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('KyivStar','Ukraine',255,'ua',380,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telesystems Of Ukraine CJSC (TSU)','Ukraine',255,'ua',380,'21');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TriMob LLC','Ukraine',255,'ua',380,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UMC/MTS','Ukraine',255,'ua',380,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Beeline','Ukraine',255,'ua',380,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('UMC/MTS','Ukraine',255,'ua',380,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Beeline','Ukraine',255,'ua',380,'68');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DU','United Arab Emirates',424,'ae',971,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Etisalat','United Arab Emirates',431,'ae',971,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Etisalat','United Arab Emirates',430,'ae',971,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Etisalat','United Arab Emirates',424,'ae',971,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Vodafone','United Kingdom',234,'gb',44,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BT Group','United Kingdom',234,'gb',44,'77');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('BT Group','United Kingdom',234,'gb',44,'76');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable and Wireless','United Kingdom',234,'gb',44,'92');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable and Wireless','United Kingdom',234,'gb',44,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable and Wireless Isle of Man','United Kingdom',234,'gb',44,'36');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cloud9/wire9 Tel.','United Kingdom',234,'gb',44,'18');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Everyth. Ev.wh.','United Kingdom',235,'gb',44,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('FlexTel','United Kingdom',234,'gb',44,'17');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Guernsey Telecoms','United Kingdom',234,'gb',44,'55');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HaySystems','United Kingdom',234,'gb',44,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G Hutchinson','United Kingdom',234,'gb',44,'20');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('H3G Hutchinson','United Kingdom',234,'gb',44,'94');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Inquam Telecom Ltd','United Kingdom',234,'gb',44,'75');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Jersey Telecom','United Kingdom',234,'gb',44,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('JSC Ingenicum','United Kingdom',234,'gb',44,'35');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lycamobile','United Kingdom',234,'gb',44,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Manx Telecom','United Kingdom',234,'gb',44,'58');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mapesbury C. Ltd','United Kingdom',234,'gb',44,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Marthon Telecom','United Kingdom',234,'gb',44,'28');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O2 Ltd.','United Kingdom',234,'gb',44,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O2 Ltd.','United Kingdom',234,'gb',44,'11');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('O2 Ltd.','United Kingdom',234,'gb',44,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('OnePhone','United Kingdom',234,'gb',44,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Opal Telecom','United Kingdom',234,'gb',44,'16');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Everyth. Ev.wh./Orange','United Kingdom',234,'gb',44,'34');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Everyth. Ev.wh./Orange','United Kingdom',234,'gb',44,'33');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PMN/Teleware','United Kingdom',234,'gb',44,'19');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Railtrack Plc','United Kingdom',234,'gb',44,'12');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Routotelecom','United Kingdom',234,'gb',44,'22');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sky UK Limited','United Kingdom',234,'gb',44,'57');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Stour Marine','United Kingdom',234,'gb',44,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Synectiv Ltd.','United Kingdom',234,'gb',44,'37');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Everyth. Ev.wh./T-Mobile','United Kingdom',234,'gb',44,'31');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Everyth. Ev.wh./T-Mobile','United Kingdom',234,'gb',44,'30');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Everyth. Ev.wh./T-Mobile','United Kingdom',234,'gb',44,'32');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','United Kingdom',234,'gb',44,'27');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Tismi','United Kingdom',234,'gb',44,'09');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Truphone','United Kingdom',234,'gb',44,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Jersey Telecom','United Kingdom',234,'gb',44,'51');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vectofone Mobile Wifi','United Kingdom',234,'gb',44,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','United Kingdom',234,'gb',44,'91');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vodafone','United Kingdom',234,'gb',44,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wave Telecom Ltd','United Kingdom',234,'gb',44,'78');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'880');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Aeris Comm. Inc.','United States',310,'us',1,'850');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'640');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel Wireless LLC','United States',310,'us',1,'510');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unknown','United States',310,'us',1,'190');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Allied Wireless Communications Corporation','United States',312,'us',1,'090');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'130');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Arctic Slope Telephone Association Cooperative Inc.','United States',310,'us',1,'710');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T Wireless Inc.','United States',310,'us',1,'150');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T Wireless Inc.','United States',310,'us',1,'680');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T Wireless Inc.','United States',310,'us',1,'070');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T Wireless Inc.','United States',310,'us',1,'560');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T Wireless Inc.','United States',310,'us',1,'410');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T Wireless Inc.','United States',310,'us',1,'380');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T Wireless Inc.','United States',310,'us',1,'170');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('AT&T Wireless Inc.','United States',310,'us',1,'980');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bluegrass Wireless LLC','United States',311,'us',1,'810');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bluegrass Wireless LLC','United States',311,'us',1,'800');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bluegrass Wireless LLC','United States',311,'us',1,'440');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cable & Communications Corp.','United States',310,'us',1,'900');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('California RSA No. 3 Limited Partnership','United States',311,'us',1,'590');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cambridge Telephone Company Inc.','United States',311,'us',1,'500');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Caprock Cellular Ltd.','United States',310,'us',1,'830');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',310,'us',1,'590');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'282');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'487');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'271');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'287');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'276');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'481');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',310,'us',1,'013');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'281');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'486');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'270');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'286');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'275');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'480');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',310,'us',1,'012');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'280');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'485');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'110');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'285');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'274');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'390');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',310,'us',1,'010');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'279');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'484');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',310,'us',1,'910');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'284');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'489');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'273');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'289');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',310,'us',1,'004');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'278');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'483');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',310,'us',1,'890');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'283');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'488');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'272');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'288');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'277');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Verizon Wireless','United States',311,'us',1,'482');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cellular Network Partnership LLC','United States',312,'us',1,'280');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cellular Network Partnership LLC','United States',312,'us',1,'270');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cellular Network Partnership LLC','United States',310,'us',1,'360');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'190');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'030');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Choice Phone LLC','United States',311,'us',1,'120');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Choice Phone LLC','United States',310,'us',1,'480');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'630');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cincinnati Bell Wireless LLC','United States',310,'us',1,'420');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cingular Wireless','United States',310,'us',1,'180');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Coleman County Telco /Trans TX','United States',310,'us',1,'620');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'040');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Consolidated Telcom','United States',310,'us',1,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Consolidated Telcom','United States',310,'us',1,'60');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',312,'us',1,'380');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'930');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'240');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'080');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cross Valliant Cellular Partnership','United States',310,'us',1,'700');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cross Wireless Telephone Co.','United States',311,'us',1,'140');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Cross Wireless Telephone Co.','United States',312,'us',1,'030');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'520');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Custer Telephone Cooperative Inc.','United States',312,'us',1,'040');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Dobson Cellular Systems','United States',310,'us',1,'440');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('E.N.M.R. Telephone Coop.','United States',310,'us',1,'990');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('East Kentucky Network LLC','United States',312,'us',1,'120');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('East Kentucky Network LLC','United States',310,'us',1,'750');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('East Kentucky Network LLC','United States',312,'us',1,'130');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Edge Wireless LLC','United States',310,'us',1,'090');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Elkhart TelCo. / Epic Touch Co.','United States',310,'us',1,'610');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'210');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Farmers','United States',311,'us',1,'311');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Fisher Wireless Services Inc.','United States',311,'us',1,'460');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GCI Communication Corp.','United States',311,'us',1,'370');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('GCI Communication Corp.','United States',310,'us',1,'430');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Get Mobile Inc.','United States',310,'us',1,'920');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'970');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Illinois Valley Cellular RSA 2 Partnership','United States',311,'us',1,'340');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'030');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Iowa RSA No. 2 Limited Partnership','United States',312,'us',1,'170');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Iowa RSA No. 2 Limited Partnership','United States',311,'us',1,'410');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Iowa Wireless Services LLC','United States',310,'us',1,'770');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Jasper','United States',310,'us',1,'650');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Kaplan Telephone Company Inc.','United States',310,'us',1,'870');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Keystone Wireless LLC','United States',312,'us',1,'180');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Keystone Wireless LLC','United States',310,'us',1,'690');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Lamar County Cellular','United States',311,'us',1,'310');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Leap Wireless International Inc.','United States',310,'us',1,'016');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'090');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Matanuska Tel. Assn. Inc.','United States',310,'us',1,'040');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Message Express Co. / Airlink PCS','United States',310,'us',1,'780');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'660');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Michigan Wireless LLC','United States',311,'us',1,'330');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'000');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Minnesota South. Wirel. Co. / Hickory','United States',310,'us',1,'400');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Missouri RSA No 5 Partnership','United States',312,'us',1,'220');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Missouri RSA No 5 Partnership','United States',312,'us',1,'010');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Missouri RSA No 5 Partnership','United States',311,'us',1,'920');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Missouri RSA No 5 Partnership','United States',311,'us',1,'020');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Missouri RSA No 5 Partnership','United States',311,'us',1,'010');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mohave Cellular LP','United States',310,'us',1,'350');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTPCS LLC','United States',310,'us',1,'570');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('NEP Cellcorp Inc.','United States',310,'us',1,'290');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nevada Wireless LLC','United States',310,'us',1,'34');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'380');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('New-Cell Inc.','United States',310,'us',1,'600');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'100');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Nexus Communications Inc.','United States',311,'us',1,'300');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('North Carolina RSA 3 Cellular Tel. Co.','United States',310,'us',1,'130');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('North Dakota Network Company','United States',312,'us',1,'230');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('North Dakota Network Company','United States',311,'us',1,'610');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Northeast Colorado Cellular Inc.','United States',310,'us',1,'450');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Northeast Wireless Networks LLC','United States',311,'us',1,'710');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Northstar','United States',310,'us',1,'670');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Northstar','United States',310,'us',1,'011');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Northwest Missouri Cellular Limited Partnership','United States',311,'us',1,'420');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'540');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Panhandle Telephone Cooperative Inc.','United States',310,'us',1,'760');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PCS ONE','United States',310,'us',1,'580');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('PetroCom','United States',311,'us',1,'170');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Pine Belt Cellular, Inc.','United States',311,'us',1,'670');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'080');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',310,'us',1,'790');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Plateau Telecommunications Inc.','United States',310,'us',1,'100');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Poka Lambro Telco Ltd.','United States',310,'us',1,'940');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'730');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'540');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Public Service Cellular Inc.','United States',310,'us',1,'500');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('RSA 1 Limited Partnership','United States',311,'us',1,'430');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('RSA 1 Limited Partnership','United States',312,'us',1,'160');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sagebrush Cellular Inc.','United States',311,'us',1,'350');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'910');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SIMMETRY','United States',310,'us',1,'46');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SLO Cellular Inc / Cellular One of San Luis','United States',311,'us',1,'260');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Smith Bagley Inc.','United States',310,'us',1,'320');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unknown','United States',310,'us',1,'15');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Southern Communications Services Inc.','United States',316,'us',1,'011');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sprint Spectrum','United States',312,'us',1,'530');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sprint Spectrum','United States',311,'us',1,'870');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sprint Spectrum','United States',311,'us',1,'490');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sprint Spectrum','United States',310,'us',1,'120');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sprint Spectrum','United States',316,'us',1,'010');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sprint Spectrum','United States',312,'us',1,'190');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sprint Spectrum','United States',311,'us',1,'880');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'210');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'260');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'200');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'250');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'160');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'240');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'660');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'230');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'31');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'220');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'270');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'280');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'330');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'800');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'300');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('T-Mobile','United States',310,'us',1,'310');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'740');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telemetrix Inc.','United States',310,'us',1,'740');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Testing','United States',310,'us',1,'14');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unknown','United States',310,'us',1,'950');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Texas RSA 15B2 Limited Partnership','United States',310,'us',1,'860');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Thumb Cellular Limited Partnership','United States',311,'us',1,'830');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Thumb Cellular Limited Partnership','United States',311,'us',1,'050');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('TMP Corporation','United States',310,'us',1,'460');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Triton PCS','United States',310,'us',1,'490');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Uintah Basin Electronics Telecommunications Inc.','United States',310,'us',1,'960');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Uintah Basin Electronics Telecommunications Inc.','United States',312,'us',1,'290');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Uintah Basin Electronics Telecommunications Inc.','United States',311,'us',1,'860');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Union Telephone Co.','United States',310,'us',1,'020');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('United States Cellular Corp.','United States',311,'us',1,'220');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('United States Cellular Corp.','United States',310,'us',1,'730');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('United Wireless Communications Inc.','United States',311,'us',1,'650');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('USA 3650 AT&T','United States',310,'us',1,'38');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VeriSign','United States',310,'us',1,'520');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unknown','United States',310,'us',1,'003');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unknown','United States',310,'us',1,'23');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unknown','United States',310,'us',1,'24');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unknown','United States',310,'us',1,'25');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('West Virginia Wireless','United States',310,'us',1,'530');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Unknown','United States',310,'us',1,'26');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Westlink Communications, LLC','United States',310,'us',1,'340');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES (NULL,'United States',311,'us',1,'150');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Wisconsin RSA #7 Limited Partnership','United States',311,'us',1,'070');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Yorkville Telephone Cooperative','United States',310,'us',1,'390');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ancel/Antel','Uruguay',748,'uy',598,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ancel/Antel','Uruguay',748,'uy',598,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Claro/AM Wireless','Uruguay',748,'uy',598,'10');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MOVISTAR','Uruguay',748,'uy',598,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Bee Line/Unitel','Uzbekistan',434,'uz',998,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Buztel','Uzbekistan',434,'uz',998,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTS/Uzdunrobita','Uzbekistan',434,'uz',998,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Ucell/Coscom','Uzbekistan',434,'uz',998,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Uzmacom','Uzbekistan',434,'uz',998,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiCel','Vanuatu',541,'vu',678,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('SMILE','Vanuatu',541,'vu',678,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiTel C.A.','Venezuela',734,'ve',58,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiTel C.A.','Venezuela',734,'ve',58,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('DigiTel C.A.','Venezuela',734,'ve',58,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movilnet C.A.','Venezuela',734,'ve',58,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Movistar/TelCel','Venezuela',734,'ve',58,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Beeline','Viet Nam',452,'vn',84,'07');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Mobifone','Viet Nam',452,'vn',84,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('S-Fone/Telecom','Viet Nam',452,'vn',84,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('VietnaMobile','Viet Nam',452,'vn',84,'05');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Viettel Mobile','Viet Nam',452,'vn',84,'08');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Viettel Mobile','Viet Nam',452,'vn',84,'06');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Viettel Mobile','Viet Nam',452,'vn',84,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Vinaphone','Viet Nam',452,'vn',84,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Digicel','Virgin Islands, U.S.',376,'vi',1340,'50');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('HITS/Y Unitel','Yemen',421,'ye',967,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Spacetel','Yemen',421,'ye',967,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Sabaphone','Yemen',421,'ye',967,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Yemen Mob. CDMA','Yemen',421,'ye',967,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Zamtel/Cell Z/MTS','Zambia',645,'zm',260,'03');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('MTN/Telecel','Zambia',645,'zm',260,'02');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Airtel/Zain/Celtel','Zambia',645,'zm',260,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Econet','Zimbabwe',648,'zw',263,'04');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Net One','Zimbabwe',648,'zw',263,'01');
    INSERT INTO mytable(Operator_Name,country,mcc,Country_ID,country_code,mnc) VALUES ('Telecel','Zimbabwe',648,'zw',263,'03');`
    inserts := strings.Split(sqlInserts, ";")
    for _, insert := range inserts {
        if insert != ""{
            stmt, _ = db.Prepare(insert)
            _, _ = stmt.Exec()
        }
    }
}
