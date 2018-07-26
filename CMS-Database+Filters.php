AB GROUPS
[Nik-United States]
order = 3
filter = ([
    country : us
    ab-group: nik
])
webapps = ({
    facebook
})

OPERATORS
order = 2
filter = ([
    operator : 666666

])
webapps = ({
    facebook
})


COUNTRIES

[United States]
order = 1
filter = ([
    country : us
])
webapps = ({

})


[Poland]
order = 0
filter = ([
    country : pl
])
webapps = ({
    facebook
})




[777777]

; ========================  Android GO  ======================================
[android_go_with_freebasic]
order = 11
inherit = ({
    "default",
    "freebasic"
})
filter = ([
    "product": "com.samsung.max.go"
    "operator" : "abc"

])
webapps = ({
    ([
        "id": "facebook",
        "rank": 1,
        "name": "Facebook",
        "homeUrl": "https://m.facebook.com/?ref=s_max_bookmark",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenUI" : ({
            "folder"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "nativeApps": ({
            "com.facebook.katana"
        }),
        "iconUrl": "ultra_apps/facebook_ultra_color_48.png"
    ]),
    ([
        "id": "makemytrip",
        "rank": 10,
        "name": "MakeMyTrip",
        "homeUrl": "https://makemytrip.com",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "iconUrl": "ultra_apps/ic_makemytrip_ultra_48.png"
    ]),
    ([
        "id": "amazon",
        "rank": 11,
        "name": "Amazon",
        "homeUrl": "https://amazon.in",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "iconUrl": "ultra_apps/ic_amazon_ultra_48.png"
    ]),
    ([
        "id": "dailyhunt",
        "rank": 12,
        "name": "DailyHunt",
        "homeUrl": "https://m.dailyhunt.in",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "iconUrl": "ultra_apps/ic_dailyhunt_ultra_48.png"
    ]),
    ([
        "id": "paytmmall",
        "rank": 13,
        "name": "Paytm Mall",
        "homeUrl": "https://paytmmall.com",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "iconUrl": "ultra_apps/ic_paytmmall_ultra_48.png"
    ]),
    ([
        "id": "nitrostreet",
        "rank": 20,
        "name": "Nitro StreetRun 2",
        "homeUrl": "http://play.ludigames.com/games/nitroStreetRun2Free/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH68",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_nitrostreet_ultra_48.png"
    ]),
    ([
        "id": "puzzlepets",
        "rank": 21,
        "name": "Puzzle Pets Pairs",
        "homeUrl": "http://play.ludigames.com/games/puzzlePetsPairsFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH68",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_puzzlepets_ultra_48.png"
    ]),
    ([
        "id": "realfootball",
        "rank": 22,
        "name": "Real Football Runner",
        "homeUrl": "http://play.ludigames.com/games/realFootballRunnerFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH68",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_realfootball_ultra_48.png"
    ]),
    ([
        "id": "ludibubbles",
        "rank": 23,
        "name": "Ludibubbles",
        "homeUrl": "http://play.ludigames.com/games/ludibubblesFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH68",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_ludibubbles_ultra_48.png"
    ]),
    ([
        "id": "duckduckgo",
        "rank": 24,
        "name": "DuckDuckGo",
        "homeUrl": "https://duckduckgo.com",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_duckduckgo_ultra_48.png"
    ]),
    ([
        "id": "worldreader",
        "rank": 25,
        "name": "Worldreader",
        "homeUrl": "https://www.worldreader.org",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_worldreader_ultra_48.png"
    ]),
})


[android_go]
order = 10
inherit = "default"
filter = ([
    "product": "com.samsung.max.go"
])
webapps = ({
    ([
        "id": "facebook",
        "rank": 1,
        "name": "Facebook",
        "homeUrl": "https://m.facebook.com/?ref=s_max_bookmark",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenUI" : ({
            "folder"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "nativeApps": ({
            "com.facebook.katana"
        }),
        "iconUrl": "ultra_apps/facebook_ultra_color_48.png"
    ]),
    ([
        "id": "makemytrip",
        "rank": 10,
        "name": "MakeMyTrip",
        "homeUrl": "https://makemytrip.com",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "iconUrl": "ultra_apps/ic_makemytrip_ultra_48.png"
    ]),
    ([
        "id": "amazon",
        "rank": 11,
        "name": "Amazon",
        "homeUrl": "https://amazon.in",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "iconUrl": "ultra_apps/ic_amazon_ultra_48.png"
    ]),
    ([
        "id": "dailyhunt",
        "rank": 12,
        "name": "DailyHunt",
        "homeUrl": "https://m.dailyhunt.in",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "iconUrl": "ultra_apps/ic_dailyhunt_ultra_48.png"
    ]),
    ([
        "id": "paytmmall",
        "rank": 13,
        "name": "Paytm Mall",
        "homeUrl": "https://paytmmall.com",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
        }),
        "iconUrl": "ultra_apps/ic_paytmmall_ultra_48.png"
    ]),
    ([
        "id": "nitrostreet",
        "rank": 20,
        "name": "Nitro StreetRun 2",
        "homeUrl": "http://play.ludigames.com/games/nitroStreetRun2Free/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH68",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_nitrostreet_ultra_48.png"
    ]),
    ([
        "id": "puzzlepets",
        "rank": 21,
        "name": "Puzzle Pets Pairs",
        "homeUrl": "http://play.ludigames.com/games/puzzlePetsPairsFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH68",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_puzzlepets_ultra_48.png"
    ]),
    ([
        "id": "realfootball",
        "rank": 22,
        "name": "Real Football Runner",
        "homeUrl": "http://play.ludigames.com/games/realFootballRunnerFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH68",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_realfootball_ultra_48.png"
    ]),
    ([
        "id": "ludibubbles",
        "rank": 23,
        "name": "Ludibubbles",
        "homeUrl": "http://play.ludigames.com/games/ludibubblesFree/?utm_source=gameloft&utm_medium=bookmark&utm_campaign=PH68",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_ludibubbles_ultra_48.png"
    ]),
    ([
        "id": "duckduckgo",
        "rank": 24,
        "name": "DuckDuckGo",
        "homeUrl": "https://duckduckgo.com",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_duckduckgo_ultra_48.png"
    ]),
    ([
        "id": "worldreader",
        "rank": 25,
        "name": "Worldreader",
        "homeUrl": "https://www.worldreader.org",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock",
            "noImages",
        }),
        "hiddenUI" : ({
            "fab"
        }),
        "iconUrl": "ultra_apps/ic_worldreader_ultra_48.png"
    ]),
})

; ========================  Other  ======================================

[global_and_preloaded_with_freebasic]
order = 6
inherit = ({
    "default",
    "freebasic"
})
filter = ([
    "~product": "com.samsung.max.go"
])
webapps = ({
    ([
        "id": "facebook",
        "rank": 1,
        "name": "Facebook",
        "homeUrl": "https://m.facebook.com/?ref=s_max_bookmark",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock"
        }),
        "nativeApps": ({
            "com.facebook.katana"
        }),
        "iconUrl": "ultra_apps/facebook_ultra_color_48.png"
    ]),
})

[global_and_preloaded]
order = 5
inherit = "default"
filter = ([
    "~product": "com.samsung.max.go"
])
webapps = ({
    ([
        "id": "facebook",
        "rank": 1,
        "name": "Facebook",
        "homeUrl": "https://m.facebook.com/?ref=s_max_bookmark",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock"
        }),
        "nativeApps": ({
            "com.facebook.katana"
        }),
        "iconUrl": "ultra_apps/facebook_ultra_color_48.png"
    ]),
})

[freebasic]
order = 1
filter = ([
    "operator": ({
        "viettel",
        "digicel-pa",
        "telcel" :6666666 777777,
        "tigo-co",
        "viva-bo",
        "mobifone",
        "freebasics",
    }),
])
webapps = ({
    ([
        "id": "freebasics",
        "rank": 6,
        "name": "Free Basics",
        "homeUrl": "https://freebasics.com/?ref=s_max_bookmark",
        "hiddenFeatures": ({
            "savings",
            "privacy",
            "adBlock",
            "noImages"
        }),
        "hiddenUI": ({
            "splash"
        }),
        "iconUrl": "ultra_apps/free_basics_48.png"
    ]),
})


[default]
order = 0
webapps = ({
    ([
        "id": "instagram",
        "rank": 2,
        "name": "Instagram",
        "homeUrl": "https://www.instagram.com/?utm_source=samsung_max_sd",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock"
        }),
        "nativeApps": ({
            "com.instagram.android"
        }),
        "iconUrl": "ultra_apps/instagram_ultra_48.png"
    ]),
    ([
        "id": "vk",
        "rank": 4,
        "name": "VK",
        "homeUrl": "https://vk.com",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock"
        }),
        "nativeApps": ({
            "com.vkontakte.android"
        }),
        "iconUrl": "ultra_apps/vkontakte_ultra_48.png"
    ]),
    ([
        "id": "cricbuzz",
        "rank": 5,
        "name": "Cricbuzz",
        "homeUrl": "http://m.cricbuzz.com",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock"
        }),
        "nativeApps": ({
            "com.cricbuzz.android",
            "com.cricbuzz.android.vernacular"
        }),
        "iconUrl": "ultra_apps/cricbuzz_ultra_48.png"
    ]),
    ([
        "id": "wikipedia",
        "rank": 7,
        "name": "Wikipedia",
        "homeUrl": "https://www.wikipedia.org",
        "defaultEnabledFeatures": ({
            "savings",
            "privacy"
        }),
        "hiddenFeatures": ({
            "adBlock"
        }),
        "nativeApps": ({
            "org.wikipedia"
        }),
        "iconUrl": "ultra_apps/wikipedia_ultra_48.png"
    ]),
})
