
var server_url = "http://localhost:8000";
var messages_url = server_url + "/messages";
var socket = new WebSocket("ws://localhost:8000/ws");
var JWTToken = "";

socket.onopen = function(event) {
    console.log("WebSocket Connected Succesfully");
};

socket.onmessage = function(event) {
    var messagesDiv = document.getElementById("messagesDiv");
    const jsonData = JSON.parse(event.data);
    /*  types:
        -2 -> message is a JWT Token
        -1 -> it is a system message
        0 -> broadcast message
        1 -> client to client whisper message
    */
    if (jsonData.Type == -1) {  //system message
        var style = 'style="color:#b32914;"'
        messagesDiv.innerHTML += "<p "+style+"><strong>SYSTEM: " + jsonData.Message + "</strong></p>";
        messagesDiv.scrollTop = messagesDiv.scrollHeight;
    }
    else if (jsonData.Type == -2) { //Token
        JWTToken = jsonData.Message;
    }
    else if (jsonData.Type == 1) {  //p2p whisper
        var style = 'style="color:#893ba8;"'
        messagesDiv.innerHTML += "<p "+style+"><strong>" + jsonData.Message + "</strong></p>";
        messagesDiv.scrollTop = messagesDiv.scrollHeight;
    }
    else { //broadcast
        var style = 'style="color:black;"'
        messagesDiv.innerHTML += "<p "+style+"><strong>" + jsonData.PlayerID + ": " + jsonData.Message + "</strong></p>";
        messagesDiv.scrollTop = messagesDiv.scrollHeight;
    }
};

socket.onclose = function(event) {
    console.log("WebSocket connection ended");
};

$(document).ready(function(){
    $('#messageInput').keypress(function(e){
        if(e.keyCode==13)
            $('#sendButton').click();
    });
});

function sendMessage() {
    const message = $('#messageInput').val();
    $.ajax({
        url: messages_url,
        type: 'POST',
        headers: {'Token':JWTToken},
        dataType: 'html',
        contentType: 'application/json',
        data: JSON.stringify(parseMessage(message)),
        success: function(response) {
            //console.log(response)
        },
        error: function(xhr, status, error) {
            console.error('Error sending message:', error);
        }
    });
    var inputField = document.getElementById("messageInput");
    inputField.value = "";
}

function parseMessage(raw_message) {
    raw_message = raw_message.toString();

    if (raw_message.slice(0,3) === "/to"){
        //whisper
        player_id = raw_message.split(" ")[1],
        message = raw_message.slice(player_id.length + 5)    // +5 stands for /to + 2 spaces
        return {
            "Type": 1,
            "PlayerID": player_id,
            "Message" : message
        }
    }
    else if (raw_message.slice(0,4) == "/cid") {
        //change playerID
        message = raw_message.split(" ")[1]
        return {
            "Type": -1,
            "PlayerID": null,
            "Message": message
        }
    }
    else {
        return {
            "Type": 0,
            "PlayerID": null,
            "Message": raw_message
        }
    }
}