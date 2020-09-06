var dispStack = '';
var resultQueue = [
    '',
    '',
    '',
    '',
    '',
    '',
    '',
    '',
    '',
    ''
]
var source = new EventSource('/connect/');
source.onmessage = function(e) {
    console.log(JSON.parse(e.data));
    results = JSON.parse(e.data);
    for (i = 9; i >= 0; i--){
        document.getElementById('r'+i).innerHTML = results['result'+i];
    }

};

function clientMsg(){
    var msg = String(dispStack + "=" + eval(dispStack));
    console.log(msg);
    fetch('http://10.0.0.153:5555/clientMsg/', {
        method: 'POST',
        body: msg
    })
        .then(console.log(msg))
    dispStack = eval(dispStack);
    document.getElementById('screen').innerHTML = dispStack;
}
function push(s){
    if (dispStack =='0'){
        dispStack = s;
    } else {
        dispStack += s;
    }
    document.getElementById('screen').innerHTML = dispStack.substring(0,9);
}
function pop(){
    dispStack = dispStack.slice(0, dispStack.length - 1);
    console.log("'"+dispStack+"'");
    if (dispStack == ''){
        dispStack = '0';
    };
    document.getElementById('screen').innerHTML = dispStack.substring(0,9);
}
function cls(){
    dispStack = '';
    document.getElementById('screen').innerHTML = '0';
}