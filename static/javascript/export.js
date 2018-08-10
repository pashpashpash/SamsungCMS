var server_post = new postRequest();
var post_url = "/post/";
function downloadConfigurations() {
    console.log("downloadConfigurations – download button pressed!");
    document.getElementsByClassName("exportLoading")[0].classList.toggle("hidden");

    var postRequestText = '{"functionToCall" : "updateConfigurationINI", "data" : {'
        +'}}';
    var postRequestJSON = JSON.parse(postRequestText);
    console.log(postRequestJSON);
    server_post.post(post_url, postRequestJSON, function(result) {
        console.log(result);
        document.getElementsByClassName("exportLoading")[0].classList.toggle("hidden");
        if(result === "success"){
            var links = [
              '/configuration.zip'
            ];
            downloadAll(links);
        }
    });
}

function downloadAll(urls) {
  var link = document.createElement('a');

  link.setAttribute('download', urls[0].substring(1));
  link.style.display = 'none';

  document.body.appendChild(link);

  for (var i = 0; i < urls.length; i++) {
    link.setAttribute('href', urls[i]);
    link.click();

  }

  document.body.removeChild(link);
}
