$(document).ready(function(){

	var oldURL = document.referrer;

	if (oldURL.split("/").pop().indexOf("login") == -1)
		$("#successLogin").alert('close');
	else
	{
		$("#successLogin").fadeTo(2000, 500).slideUp(500, function(){
    		$(this).alert('close');
		});
	}

	UpdateLobby();
	var myInterval = setInterval(UpdateLobby, 5000);

	function UpdateLobby()
	{
		$.ajax({
			url: "/lobby/games",
			type: "POST",
			async: false
		}).done(function(resp)
		{
			resp = JSON.parse(resp);

			//console.log(resp);

			var myList = "";
			for (var i = 0; i < resp.length; i++)
			{
				myList += "<tr class='success'>"
				var p = (resp[i].Players[0] == "" ? resp[i].Players[1] : resp[i].Players[0]);
				if (resp[i].Joinable == 1)
					myList += "<td><a href='/game/joinGame?id="+resp[i].GameId+"'> <strong>Join game:</strong> "+p+"</a></td>";
				else
					myList += "<td> Game: "+resp[i].Players[0]+" vs "+resp[i].Players[1]+"</td>";
				myList += "</tr>";
			}

			$("#listGames").html(myList);
		});
	}

});