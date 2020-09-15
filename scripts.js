let dispStack = '';
let source = new EventSource('/connect/');
let userID = window.prompt("Please enter your name.");
let updating = false;
let updateID = 0;
let url = "http://calcrudlator.herokuapp.com";
//let url = "http://localhost:8080";

source.onmessage = function(e) {
    readMsg(e);
}
function routeCreateUpdate(){
    if (!updating){
        createMsg()
    } else {
        updateMsg()
    }
}
function prepareUpdate(rowID){
    updating = true;
    updateID = parseInt(document.getElementById('row'+rowID+'messageID').innerHTML,10);
    dispStack = document.getElementById('row'+rowID+'entry').innerHTML.split('=')[0];
    document.getElementById('screen').innerHTML = dispStack;
}
function createMsg(){
    let result = String(eval(dispStack)).substr(0, 12)
    let entry = String(dispStack + "=" + result);

    fetch(url+'/api/entry/', {
        method: 'POST',
        body: JSON.stringify({
            userID: userID,
            entry: entry
        }),
        headers: {
            'Content-Type': 'application/json'
        }
    });

    dispStack = result;
    document.getElementById('screen').innerHTML = dispStack;
}
function readMsg(e){
    console.log(JSON.parse(e.data));
    let data = JSON.parse(e.data);
    for (let i=0;i<=9;i++){
        document.getElementById('row'+i+'messageID').innerHTML = data[i]['messageID'];
        document.getElementById('row'+i+'timestamp').innerHTML = data[i]['timestamp'].substring(11,16);
        document.getElementById('row'+i+'userID').innerHTML = data[i]['userID'];
        document.getElementById('row'+i+'entry').innerHTML = data[i]['entry'];
    }
}
function updateMsg(rowID){
    let result = String(eval(dispStack)).substr(0, 12)
    let entry = String(dispStack + "=" + result);
    fetch(url+'/api/entry/'+rowID, {
        method: 'PUT',
        body: JSON.stringify({
            messageID: updateID,
            userID: userID,
            entry: entry
        }),
        headers: {
            "Content-Type": "application/json"
        }
    });
    updating = false;
    updateID = 0;
}
function deleteMsg(rowID){
    console.log(document.getElementById('row'+rowID+'messageID').innerHTML)
    fetch(url+'/api/entry/'+rowID, {
        method: 'DELETE',
        body: JSON.stringify({
            messageID: parseInt(document.getElementById('row'+rowID+'messageID').innerHTML,10)
        }),
        headers: {
            "Content-Type": "application/json"
        }
    });
}
function push(s) {
    // For a straightforward 'backspace' functionality, implement a stack and later pop() to delete last character
    if (dispStack == '0') {
        dispStack = s;
    } else if (dispStack.length == 12){
        // do nothing
        console.log("Do nothing");
    } else {
        dispStack += s;
    }
    document.getElementById('screen').innerHTML = dispStack;
}
function pop(){
    dispStack = dispStack.slice(0, dispStack.length - 1);
    console.log("'"+dispStack+"'");
    if (dispStack == ''){
        dispStack = '0';
    }
    document.getElementById('screen').innerHTML = dispStack.substring(0,9);
}
function cls(){
    dispStack = '';
    updating = false;
    updateID = 0;
    document.getElementById('screen').innerHTML = '0';
}