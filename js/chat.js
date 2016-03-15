$('document').ready(function()
{
    /***** Variables *****/
    var msgSound = new Audio();    
    var conn;
    var msg = $("#chat #msg");
    var log = $("#chat #log");    
    var chatbody =  $('#chat .panel-body'); 
    var chattitle = $('#chat .panel-title');
    var text = " <p class='newmsg'> - <b> New Message </b></p>";
    var c = false;
    /***** END Variables *****/


    /***** Main *****/
    msgSound.src = '../sounds/blop.mp3';

    if (window["WebSocket"]) 
    {
        conn = new WebSocket("ws://" + host + "/ws");
        conn.onclose = function(evt) 
        {
            appendLog($("<div><b>Connection closed.</b></div>"));          
        }
        conn.onmessage = function(evt) 
        {                
            var command = evt.data.substring(0, 5);
            if (command == "\\move" || command == "\\join" || command == "\\exit")
            {
                updateBoard();
                return;
            }


            appendLog($("<p/>").text(evt.data));
            msgSound.play();
            if( chatbody.css('display') == 'none' && !c)
            {
                c = true;
                chattitle.append(text);
            }
        }
    } 
    else 
    {
        appendLog($("<div><b>Your browser does not support WebSockets.</b></div>"));
    }
    /***** END Main *****/

    
    /***** Functions *****/
    function appendLog(msg) 
    {
        var d = log[0]
        var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
        msg.appendTo(log)
        if (doScroll) 
        {
            d.scrollTop = d.scrollHeight - d.clientHeight;
        }
    }
    /***** END Functions *****/


    /***** Events *****/
    msg.focus(function()
    {     
        $('.newmsg').hide();
        c = false;
    });

    $("#form").submit(function() 
    {
        if (!conn) 
        {
            return false;
        }
        if (!msg.val()) 
        {
            return false;
        }
        conn.send(msg.val());
        appendLog($("<div><p id='umsg'>" + msg.val() + "</p></div>"));
        msg.val("");
        return false
    });
    /***** END Events *****/
    
});