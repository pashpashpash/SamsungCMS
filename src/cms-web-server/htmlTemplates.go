package main
func configPageHTML(Config_ID string) (string){
    return (`<!DOCTYPE html><html><head>
                <link rel="stylesheet" type="text/css" href="/stylesheets/main.css">
        </head>

        <body>
            <div id="header">
                <div id="headerIcon" onclick="location.reload();location.href='../index.html'"></div>
                <div id="headerText" onclick="location.reload();location.href='../index.html'"> Ultra Configuration <span id="smallerText">#`+Config_ID+`</span></div>
            </div>
            <hr>



            <!-- Javascript Includes -->

            <script type="text/javascript" src="../javascript/rest.js"></script>
            <script type="text/javascript" src="../javascript/post.js"></script>
            <script type="text/javascript"> var imported_config_id = "`+Config_ID+`"</script>
            <script type="text/javascript" src="../javascript/configs.js"></script>

    </body></html>`)
}

func exportPageHTML(Config_ID string) (string){
    return (`<!DOCTYPE html><html><head>
                <link rel="stylesheet" type="text/css" href="/stylesheets/main.css">
        </head>

        <body>
            <div id="header">
                <div id="headerIcon" onclick="location.reload();location.href='../index.html'"></div>
                <div id="headerText" onclick="location.reload();location.href='../index.html'"> Ultra Configurations <span id="smallerText">`+`Export`+`</span></div>
            </div>
            <hr>
            <button id="download" onClick = "downloadConfigurations()">Export Ultra Configurations</button>


            <!-- Javascript Includes -->

            <script type="text/javascript" src="../javascript/rest.js"></script>
            <script type="text/javascript" src="../javascript/post.js"></script>
            <script type="text/javascript" src="../javascript/export.js"></script>

    </body></html>`)
}
