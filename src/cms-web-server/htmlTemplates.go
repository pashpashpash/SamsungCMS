package main
func appViewHTML(appName string) (string){
    return `
 <!DOCTYPE html>
 <html>
 <head>
       <title>` + appName +`</title>
       <link rel="stylesheet" type="text/css" href="../stylesheets/main.css">
 </head>
 <body>
     <div id = "header">
         <div id="headerIcon" onClick="location.reload();location.href='../index.html'"></div><div id="headerText" onClick="location.reload();location.href='../index.html'"> Ultra Apps <span id="smallerText">CMS</span></div>
     </div>
     <div id = "filters">
     <div id = "filterContainer">
         <div id="filterText">
             Country
         </div>
         <select name="country"  onchange="selectChange(this); this.oldvalue = this.value;">
             <option value="star">🔯</option>
           <option value="af">Afghanistan</option>

           <option value="ge">Germany</option>
           <option value="uk">United Kingdom</option>
         </select>
     </div>
     <div id = "filterContainer">
         <div id="filterText">
             Operator
         </div>
         <select name="operator" onchange="selectChange(this); this.oldvalue = this.value;">
               <option value="star">🔯</option>
           <option value="tmobile(333444)">T-Mobile (333444)</option>

           <option value="T-Mobile(333445)">T-Mobile (333445)</option>
           <option value="AT&T(456932)">AT&T (456932)</option>
         </select>
     </div>
     <div id = "filterContainer">
         <div id="filterText">
             Version
         </div>
         <select name="version"  onchange="selectChange(this); this.oldvalue = this.value;">
             <option value="star">🔯</option>
             <option value="2.4">2.4</option>
           <option value="2.5">2.5</option>
           <option value="2.6">2.6</option>
         </select>
     </div>
         <div id = "starandsearch">
   <div class="clicked" id="star">
   </div>
    <input class="search" type="text" placeholder="Search..">
         </div>
     </div>
     <div class ="webApp">
     </div>
     </main>
     <script type="text/javascript" src="../javascript/filters.js"></script>
     <script type="text/javascript" src="../javascript/rest.js"></script>
     <script type="text/javascript" src="../javascript/appView.js"></script>
 </body>
 `
}
