$(document).ready(function(){

	/***** Variables *****/
	var oldURL = document.referrer;
	var lobbyInterval;
	/***** END Variables *****/

	/***** Main *****/
	if (oldURL.split("/").pop().indexOf("login") == -1)
		$("#successLogin").alert('close');
	else
	{
		$("#successLogin").fadeTo(2000, 500).slideUp(500, function(){
    		$(this).alert('close');
		});
	}

	UpdateLobby();
	lobbyInterval = setInterval(UpdateLobby, 1000);
	/***** END Main *****/


	/***** Functions *****/
	function UpdateLobby()
	{
		$.ajax({
			url: "/lobby/games",
			type: "POST"
		}).done(function(resp)
		{
			resp = JSON.parse(resp);

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

		$.ajax({
			url: "/lobby/players",
			type: "POST"
		}).done(function(resp)
		{
			resp = JSON.parse(resp);

			var myList = "";
			for (var i = 0; i < resp.length; i++)
			{
				myList += "<tr class='info'>";
				myList += "<td><a name='player'>"+resp[i].Username+"</a></td>";
				myList += "</tr>";
			}

			$("#listPlayers").html(myList);
		});
	}
	/***** END Functions *****/


	/***** Events *****/
	$("body").on("click", "[name='player']", function()
	{
		var playerUsername = $(this).html();
		$("#modal-player .modal-title").html(playerUsername);
		$("#modal-player").modal();

		if (username == playerUsername)
			$("#btn-challenge").hide();
		else
			$("#btn-challenge").show();
	});

	$("#btn-challenge").click(function()
	{
		var playerUsername = $("#modal-player .modal-title").html();
		clearInterval(lobbyInterval);

		$.ajax({
			url: "/lobby/challenge",
			type: "POST",
			data: {rival: playerUsername}
		}).done(function(){
			window.location.href = "/game";
		});
	});
	/***** END Events *****/
});