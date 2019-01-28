var OBJECTS = {
    GRAPH_SVG_ID: "#pricegraph-svg",
    POINT_A_CRCL_ID: "pointACircle",
    POINT_B_CRCL_ID: "pointBCircle"
};

var LastPoint = {};
var PointSelection = [];

function updatePointsUI(){
    switch (PointSelection.length) {
        case 0:{
            document.getElementById("pointA").innerHTML = "A: -";
            document.getElementById("pointB").innerHTML = "B: -";
            break;
        } case 1: {
            document.getElementById("pointA").innerHTML = "A: " + PointSelection[0].price + " | " + PointSelection[0].graphdate;
            document.getElementById("pointB").innerHTML = "B: ";
            break;
        } case 2: {
            document.getElementById("pointA").innerHTML = "A: " + PointSelection[0].price + " | " + PointSelection[0].graphdate;
            document.getElementById("pointB").innerHTML = "B: " + PointSelection[1].price + " | " + PointSelection[1].graphdate;
            break;
        }
    }
}

function selectPoint(){

    if (Object.keys(LastPoint).length === 0 && LastPoint.constructor === Object){
        return
    }

    if (PointSelection.length == 0 || PointSelection.length == 1){
        PointSelection.push(LastPoint)
    } else {
        PointSelection = [];
    }

    updatePointsUI();
}