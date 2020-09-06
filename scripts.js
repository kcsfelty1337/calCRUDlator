
var dispStack = '';


// Create a new HTML5 EventSource
var source = new EventSource('/events/');

// Create a callback for when a new message is received.
source.onmessage = function(e) {

    // Append the `data` attribute of the message to the DOM.
    document.body.innerHTML += e.data + '<br>';
};

function clientMsg(){
    //var dispStack = '1+2+3*4';
    var msg = String(dispStack + "=" + eval(dispStack));
    console.log(msg);
    fetch('http://localhost:8000/clientMsg/', {
        method: 'POST',
        body: msg
    })
        .then(console.log(msg))
}

function push(s){
    dispStack += s;
    document.getElementById('disp').innerHTML = dispStack;
}

function pop(){
    dispStack = dispStack.slice(0, dispStack.length-1);
    document.getElementById('disp').innerHTML = dispStack;
}

function cls(){
    dispStack = '';
    document.getElementById('disp').innerHTML = '0';
}