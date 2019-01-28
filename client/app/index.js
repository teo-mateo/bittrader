console.log("this and that");

var GRAPHDATA = [];
var POSITIONS = {};
const URL = {
    POSITIONS: "/positions",
    UPDATES: "/updates",
    PRICE_SETS:"/misc/price1msim",
    DUPLICATE_PRICE_SET:"/misc/price1msim/duplicate",
    CREATE_PRICE1m: "/misc/price1msim/create",
    DELETE_PRICE1m: "/misc/price1msim/delete"
};

const VZOOM = 15;
const BOXWIDTH = 100;

var f = function(n) { return n;}
//f = Math.log2

function d3init() {

    var w = 500;
    var h = 80;

    var svg = d3.select("#positions-div")
        .append("svg")
        .attr("id", "svg-positions")
        .attr("width", w)
        .attr("height", h)
        .attr("style", "border: 1px solid black;");



    function positiony(pos, i){
        var hu = 0;
        for (var j = 0; j <i; j++) {
            hu += (POSITIONS[j].high - POSITIONS[j].low);
        }
        return f(hu) * VZOOM;
    }

    d3.json(URL.POSITIONS, function (json) {
        POSITIONS = json;
        svg.selectAll("rect")
            .data(POSITIONS)
            .enter()
            .append("rect")
                .attr("class", "position-box")
                .attr("id", function(pos) { return "pos-" + pos.nrcrt; })
                .attr("x", 0)
                .attr("data-nrcrt", function(pos) { return pos.nrcrt; })
                .attr("y", positiony)
                .attr("height", function(pos, i) {
                    console.log(i);
                    var h = pos.high - pos.low;
                    return f(h) * VZOOM;
                })
                .attr("width", BOXWIDTH)
                .attr("transform", "translate(0, 0)");

        //resize positions svg
        var height = 0, width = 0;
        d3.selectAll(".position-box").each(function(rect){
            height += +d3.select(this).attr("height");
        });
        console.log("svg height: " + height);
        d3.select("#svg-positions").attr("height", height);


        svg.selectAll("rect.pos-fiatbox")
            .data(POSITIONS)
            .enter()
            .append("rect")
            .attr("class", "pos-fiatbox")
            .attr("id", function(pos) { return "pos-fiatbox-" + pos.nrcrt;})
            .attr("x", BOXWIDTH)
            .attr("y", positiony)
            .attr("height", 3)
            .attr("width", function(pos, i) {
                return pos.fiat;
            })
            .attr("fill", "green");

        svg.selectAll("rect.pos-cryptobox")
            .data(POSITIONS)
            .enter()
            .append("rect")
            .attr("class", "pos-cryptobox")
            .attr("id", function(pos) { return "pos-cryptobox-" + pos.nrcrt;})
            .attr("x", BOXWIDTH)
            .attr("y", function(pos, i) {
                return positiony(pos, i) + 4;
            })
            .attr("height", 3)
            .attr("width", function(pos, i) {
                return pos.crypto * pos.low;
            })
            .attr("fill", "orange");

        svg.selectAll("rect.pos-hitsbox")
            .data(POSITIONS)
            .enter()
            .append("rect")
            .attr("class", "pos-hitsbox")
            .attr("id", function(pos) { return "pos-hitsbox-" + pos.nrcrt;})
            .attr("x", BOXWIDTH)
            .attr("y", function(pos, i) {
                return positiony(pos, i) + 6;
            })
            .attr("height", 3)
            .attr("width", function(pos) {
                return pos.hits;
            })
            .attr("fill", "grey");


        var labelsgroups = svg.selectAll(".g-boxlabel")
            .data(POSITIONS)
            .enter()
            .append("g")
            .attr("class", "g-boxlabel")
            .attr("id", function(pos) { return "ubound-"+pos.nrcrt })
            .attr("transform", function(pos, i) {
                var hu = 0;
                for (var j = 0; j <i; j++) {
                    hu += (POSITIONS[j].high - POSITIONS[j].low);
                }
                var htranslate = f(hu) * VZOOM + 0.5*VZOOM;
                return "translate(3," + htranslate.toString() + ")";
            });

        labelsgroups
            .append("text")
            .style("font-size", (0.4*VZOOM).toString() + "px")
            .style("font-family", "sans-serif")
            .text(function(pos) {
                return "(" + pos.nrcrt + ") " + pos.high.toFixed(3) ;
            })
            .attr("dy", "-.35em")
            .attr("opacity", 50);

        labelsgroups
            .append("text")
            .style("font-size", (0.15*VZOOM).toString() + "px")
            .style("font-family", "sans-serif")
            .attr("id", function(pos) { return "hitcount-"+pos.nrcrt })
            .text(function(pos) { return pos.hitcount + " hits" })
            .attr("dy", "-.35em")
            .attr("opacity", 50)
            .attr("transform", function(pos, i) {
               return "translate(" +BOXWIDTH+", -5)";
            });

        labelsgroups
            .append("text")
            .style("font-size", (0.15*VZOOM).toString() + "px")
            .style("font-family", "sans-serif")
            .attr("id", function(pos) { return "poscrypto-"+pos.nrcrt })
            .text(function(pos) { return "CRYPTO: " + pos.crypto.toFixed(5) })
            .attr("dy", "-.35em")
            .attr("opacity", 50)
            .attr("transform", function(pos, i) {
                return "translate(" +BOXWIDTH+", -2.5)";
            });

        labelsgroups
            .append("text")
            .style("font-size", (0.15*VZOOM).toString() + "px")
            .style("font-family", "sans-serif")
            .attr("id", function(pos) { return "posfiat-"+pos.nrcrt })
            .text(function(pos) { return "FIAT: " + pos.fiat.toFixed(5) })
            .attr("dy", "-.35em")
            .attr("opacity", 50)
            .attr("transform", function(pos, i) {
                return "translate(" +BOXWIDTH+", -0.8)";
            });

    });

    //add dummy price
    var price = 190.3;
    d3.select("#svg-positions")
        .append("g")
        .attr("transform", "translate(" + (-BOXWIDTH) + "," + price + ")")
        .attr("id", "g-price")


    //dot
    d3.select("#g-price")
        .append("circle")
        .attr("cx", 0)
        .attr("cy", 0)
        .attr("r", 1)
        .attr("fill", "red");

    //price
    d3.select("#g-price")
        .append("text")
        .attr("id", "pricelabel")
        .attr("transform", "translate(5, 1)")
        .attr("opacity", 100)
        .attr("dy", "-.35em")
        .style("font-size", "8px")
        .style("font-family", "sans-serif")
        .text(price.toString())

    //line
    d3.select("#g-price")
        .append("line")
        .attr("x1", 0)
        .attr("y1", 0)
        .attr("x2", 150)
        .attr("y2", 0)
        .attr("stroke-width", 0.5)
        .attr("stroke", "red");

    //time
    d3.select("#svg-footer")
        .append("text")
        .attr("id", "pricetime")
        .attr("transform", "translate(5, 10)")
        .attr("opacity", 100)
        .attr("dy", "-.35em")
        .style("font-size", "6.5px")
        .style("font-family", "sans-serif")
        .text("2017-01-01 00:00:00");

    //profit percent
    d3.select("#svg-footer")
        .append("text")
        .attr("id", "profitpercent")
        .attr("transform", "translate(5, 20)")
        .attr("opacity", 100)
        .attr("dy", "-.35em")
        .style("font-size", "6.5px")
        .style("font-family", "sans-serif")
        .text("Profit 0%");

    //total assets
    d3.select("#svg-footer")
        .append("text")
        .attr("id", "totalassets")
        .attr("transform", "translate(5, 30)")
        .attr("opacity", 100)
        .attr("dy", "-.35em")
        .style("font-size", "6.5px")
        .style("font-family", "sans-serif")
        .text("Assets 0.0");

}

var serverSource = {};

function subscribeToBittrader(){
    //subscribe to SSE
    serverSource = new EventSource(URL.UPDATES);
    serverSource.onmessage = serverUpdate;
}

function unsubscribeToBittrader(){
    serverSource.close();
}

function serverUpdate(msg){
    {
        var msgjson = JSON.parse(msg.data);

        //console.log(msgjson);

        updatePrice(msgjson);

        if (msgjson.buy !== null){
            updatePosition(msgjson.buy.position);
        }

        if (msgjson.sell != null){
            updatePosition(msgjson.sell.position);
        }

        if (msgjson.assignprofit != null) {
            updatePosition(msgjson.assignprofit);
        }

        if(msgjson.redistribution != null) {
            //get all positions;
            axios.get(URL.POSITIONS)
                .then(function(response){
                    POSITIONS = response.data;
                    console.log("Redistribution");
                    for (var i = 0; i < POSITIONS.length; i++){
                        updatePosition(POSITIONS[i])
                    }
                })
                .catch(function(error){
                    console.log(error);
                })

            //update across the board
        }
    }
}

function updatePosition(position){
    for (var i = 0; i < POSITIONS.length; i++){
        if (position.nrcrt == POSITIONS[i].nrcrt){
            POSITIONS[i] = position;
            break;
        }
    }

    var pbox = d3.select("#positionbox-"+position.nrcrt);
    if (position.fiat > 0.0){
        pbox.attr("fill", "green");
        pbox.attr("width", position.fiat * 1000);
        console.log("fiat " + position.fiat);
    } else if (position.crypto > 0.0){
        pbox.attr("fill", "orange");
        pbox.attr("width", position.crypto * position.low * 1000);
        console.log("fiat " + position.fiat);
    }




    // var lblHits = d3.select("#hitcount-"+position.nrcrt);
    // lblHits.text(position.hitcount + " hits");
    //
    // var lblCrypto = d3.select("#poscrypto-"+position.nrcrt);
    // lblCrypto.text("CRYPTO: " + position.crypto.toFixed(5));
    //
    // var lblFiat = d3.select("#posfiat-"+position.nrcrt);
    // lblFiat.text("FIAT: " + position.fiat.toFixed(5));
    //
    // var cryptobox = d3.select("#pos-cryptobox-" + position.nrcrt);
    // cryptobox.attr("width", position.crypto * position.low);
    //
    // var fiatbox = d3.select("#pos-fiatbox-" + position.nrcrt);
    // fiatbox.attr("width", position.fiat);
    //
    // var hitsbox = d3.select("#pos-hitsbox-" + position.nrcrt);
    // hitsbox.attr("width", position.hits);
}

function updatePrice(msg){

    var price = msg.price;
    var time = msg.time;
    var top = POSITIONS[0].high;
    var bottom = POSITIONS[POSITIONS.length-1].low;
    var ytranslate = f(((top-bottom) - (price-bottom))) * VZOOM;

    d3.select("#svg-positions")
        .select("#g-price")
        .attr("transform", "translate(0, " + ytranslate+ ")");

    d3.select("#pricelabel")
        .text(price.toFixed(3));

    d3.select("#pricetime")
        .text(time);

    d3.select("#profitpercent")
        .text("Profit " + msg.profitpercent.toFixed(3) + "%");

    d3.select("#totalassets")
        .text("Assets " + msg.totalassets.toFixed(3));

    if (typeof xscale === "function" && typeof yscale === "function"){
        var t = Date.parse(time);
        var rectx = xscale(t);
        //console.log(msg.nrcrt + " " + time + " " + rectx);

        d3.select("#priceline")
            .attr("x1", rectx)
            .attr("x2", rectx);
    }
}

function priceSetsInit(){
    axios.get(URL.PRICE_SETS)
        .then(function(response){

            d3.select("#selectPriceSets").selectAll("*").remove()

            d3.select("#selectPriceSets")
                .selectAll("option")
                .data(response.data)
                .enter()
                .append("option")
                .attr("id", function(d){
                    return d.id;
                })
                .html(function(d){
                    var html = d.id + "| " + d.description + " "  + d.mintime.slice(0,10) + " - "+ d.maxtime.slice(0,10);
                    return html;
                });
        })
        .catch(function(error){
            console.log(error);
        })
}

function createPrice1mSim(){
    if (!confirm("Create prices?")){
        return
    }

    var v = document.getElementById("selectPriceSets").value;
    var ix = v.indexOf("|");
    var id = v.substring(0, ix);

    var ix2 = v.substring(ix+2).indexOf(" ");
    var newname =  v.substring(ix+2, ix+2+ix2) + "(sim)";
    var sim = {
        "simid": +id,
        "newsimname": newname,
        "newsimdesc": newname,
        "starttime": PointSelection[0].graphdate,
        "endtime": PointSelection[1].graphdate
    };

    axios.post(URL.CREATE_PRICE1m, sim)
        .then(function(){
            priceSetsInit();
        })
        .catch(function(error){
            console.log(error);
        })
}

function deletePrice1mSim(){
    if (!confirm("Delete price set?")){
        return
    }
    var v = document.getElementById("selectPriceSets").value;
    var ix = v.indexOf("|");
    var id = v.substring(0, ix);

    axios.post(URL.DELETE_PRICE1m, id)
        .then(function(){
            priceSetsInit();
        })
        .catch(function(error){
            console.log(error);
        })

}

function duplicatePrice1mSim(){

    if (!confirm("Duplicate price set?")){
        return
    }

    var v = document.getElementById("selectPriceSets").value;
    var ix = v.indexOf("|");
    var id = v.substring(0, ix);

    var ix2 = v.substring(ix+2).indexOf(" ");
    var newname =  v.substring(ix+2, ix+2+ix2) + "(sim)";
    var newsim = {
        "simid": +id,
        "newsimname": newname,
        "newsimdesc": newname
    };

    axios.post(URL.DUPLICATE_PRICE_SET, newsim)
        .then(function(){
            priceSetsInit()
        })
        .catch(function(error){
            console.log(error)
        })
}

function buildPriceSetGraph(){
    var v = document.getElementById("selectPriceSets").value;
    var ix = v.indexOf("|");
    var id = v.substring(0, ix);
    buildPriceSetGraph2(id, 2);
}

var xscale = {}, yscale = {}

function buildPriceSetGraph2(id, minuteskip){

    var overlays = d3.selectAll("rect.overlay");
    overlays
        .on("click", null)
        .on("mousemove", null)
        .on("mouseover", null);
    d3.selectAll("#pricegraph-svg").remove();

    //graph dimensions
    var svg = d3.select("#pricegraph-div")
            .append("svg")
            .attr("id", "pricegraph-svg")
            .attr("height", 500)
            .attr("width", 1000)
            .attr("style", "border: 1px solid black;"),
        margin = {top: 20, right: 20, bottom: 30, left: 50},
        width = +svg.attr("width") - margin.left - margin.right,
        height = +svg.attr("height") - margin.top - margin.bottom,
        g = svg.append("g").attr("transform", "translate(" + margin.left + ", " + margin.top + ")");


    var strictIsoParse = d3.utcParse("%Y-%m-%dT%H:%M:%SZ"),
        bisectDate = d3.bisector(function(d) { return d.graphdate; }).left,
        formatValue = d3.format(",.2f"),
        formatCurrency = function(d) { return "$" + formatValue(d); };



    //set the ranges
    xscale = d3.scaleTime().rangeRound([0, width]);
    yscale = d3.scaleLinear().rangeRound([height, 0]);

    //define the line
    var line = d3.line()
        .x(function(d) { return xscale(d.graphdate); })
        .y(function(d) { return yscale(d.price); });

    var theurl = URL.PRICE_SETS + "/" + id + "/" + minuteskip;
    axios.get(theurl)
        .then(function(response){

            GRAPHDATA = response.data;

            GRAPHDATA.forEach(function(d){
                d.graphdate = strictIsoParse(d.time);
                d.price = +d.price;
            });

            xscale.domain(d3.extent(GRAPHDATA, function(d) { return d.graphdate; }));
            yscale.domain(d3.extent(GRAPHDATA, function(d) { return d.price; }));

            g.append("g")
                .attr("transform", "translate(0," + height + ")")
                .attr("class", "x axis")
                .call(d3.axisBottom(xscale))
                .select(".domain")
                .remove();

            g.append("g")
                .call(d3.axisLeft(yscale))
                .append("text")
                .attr("class", "y axis")
                .attr("fill", "#000")
                .attr("transform", "rotate(-90)")
                .attr("y", 6)
                .attr("dy", "0.71em")
                .attr("text-anchor", "end")
                .text("Price (EUR)");

            g.append("path")
                .datum(GRAPHDATA)
                .attr("fill", "none")
                .attr("stroke", "steelblue")
                .attr("stroke-linejoin", "round")
                .attr("stroke-linecap", "round")
                .attr("stroke-width", 1.5)
                .attr("d", line);

            var focus = svg.append("g")
                //.attr("transform", "translate(" + margin.left + ", " + margin.top + ")")
                .attr("class", "focus")
                .style("display", "none");
            focus.append("circle")
                .attr("r", 4.5);
            focus.append("text")
                .attr("x", 9)
                .attr("dy", ".35em");

            var overlay = svg.append("rect")
                .attr("class", "overlay")
                .attr("width", width)
                .attr("height", height)
                .attr("transform", "translate(" + margin.left + ", " + margin.top + ")")
                .on("mouseover", function() {
                    focus.style("display", null);
                })
                .on("mouseout", function() {
                    focus.style("display", "none");
                })
                .on("mousemove", mousemove)
                .on("click", click);

            d3.select("#positions-group").remove();
            var posgroup = d3.select("#pricegraph-svg")
                .append("g")
                .attr("id", "positions-group")
                .attr("transform", "translate(55, 20)");

            //price line
            posgroup.append("line")
                .attr("id", "priceline")
                .attr("x1", 100)
                .attr("y1", margin.top)
                .attr("x2", 100)
                .attr("y2", height)
                .attr("stroke", "red")
                .attr("stroke-width", 2);

            function mousemove(){

                var x0 = xscale.invert(d3.mouse(this)[0]),
                    i = bisectDate(GRAPHDATA, x0, 1),
                    d0 = GRAPHDATA[i-1],
                    d1 = GRAPHDATA[i],
                    d = x0 - d0.graphdate > d1.graphdate - x0 ? d1 : d0;
                var translx = margin.left + xscale(d.graphdate),
                    transly = margin.top + yscale(d.price);
                focus.attr("transform", "translate(" + translx + ", " + transly + ")");

                var txt = d3.format(",.2f")(d.price) + " | " + d.graphdate.toISOString().slice(0, 19);
                focus.select("text").text(txt);

                LastPoint = d;
            }

            function click(){
                selectPoint();

                if (PointSelection.length !== 2){
                    d3.select("#selection-line").remove();
                    return;
                }

                svg
                    .append("line")
                    .attr("transform", "translate(" + margin.left + ", " + margin.top + ")")
                    .attr("id", "selection-line")
                    .attr("x1", xscale(PointSelection[0].graphdate))
                    .attr("y1", yscale(PointSelection[0].price))
                    .attr("x2", xscale(PointSelection[1].graphdate))
                    .attr("y2", yscale(PointSelection[1].price))
                    .style("stroke", "red")
                    .style("stroke-width", 1)
                    .style("display", null);
            }

        })
        .catch(function(error){
            console.log(error);
        });
}



function addPositionToGraph2(pos){

    var recty1 = yscale(pos.low);
    var recty2 = yscale(pos.high);


    var posrect = d3.select("#positions-group")

        .append("rect")
        .attr("id", "positionbox-" + pos.nrcrt)
        .style("opacity", 20)
        .attr("fill", "orange")
        .attr("height", recty1-recty2)
        .attr("stroke", "yellow").attr("stroke-width", 0.3)
        .attr("transform", "translate(0, " + recty2 + ")");


    if (pos.crypto > 0.0){
        posrect.attr("height", pos.crypto * pos.low);
    } else if (pos.fiat > 0.0){
        posrect.attr("width", pos.fiat);
    }





}

function addPositionsToGraph() {

    console.log("start addPositionsToGraph: " + POSITIONS.length + " positions");

    unsubscribeToBittrader();


    var posrect = d3.select("#positions-group");
    posrect
        .selectAll("rect")
        .data(POSITIONS)
        .enter()
        .append("rect")
        .attr("id", function(pos) { return "positionbox-" + pos.nrcrt;})
        //.style("opacity", 20)
        .attr("fill", "orange")
        .attr("height", function(pos) {
            var recty1 = yscale(pos.low);
            var recty2 = yscale(pos.high);
            return recty1-recty2;
        })
        .attr("width", function(pos){
            if (pos.crypto > 0.0){
                return pos.crypto * pos.low;
            } else if (pos.fiat > 0.0){
                return pos.fiat
            }
        })
        .attr("stroke", "yellow").attr("stroke-width", 0.3)
        .attr("transform", function(pos) {
            return "translate(0, " + yscale(pos.high) + ")"
        });


    // POSITIONS.forEach(function (pos) {
    //     addPositionToGraph2(pos);
    // });

    subscribeToBittrader();

    console.log("end addPositionsToGraph");
}