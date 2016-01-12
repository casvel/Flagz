$('document').ready(function()
{
	var r, c;
	var turn;
	var players = ["", ""];
	var score = [0, 0];
	var mines;
	var username;
	var myPos, rivalPos;

	initBoard();

	var myInterval = setInterval(updateBoard, 1000); 

	$(window).unload(function(){
		$.ajax({
			url: "/game/exit", 
			type: "POST",
			async: false
		});
	});

	function stringPlayers()
	{
		if (turn == 0)
			return "<strong style='color:#D9534F'>"+players[0]+"</strong> vs "+players[1];
		else
			return players[0]+" vs <strong style='color:#D9534F'>"+players[1]+"</strong>";
	}

	function updateHTML()
	{
		$("#players").html(stringPlayers());
		$("#myScore").html("Score: "+score[myPos]);
		$("#rivalScore").html("Score: "+score[rivalPos]);
		$("#rivalUsername").html(players[rivalPos]);
		$("#mines").html("Mines: <strong>"+mines+"</strong>");

		if (players[turn] == username)
		{
			$("#rivalBlock").css("background-color", "white");
			$("#rivalUsername").css("color", "#325D88");
			$("#myBlock").css("background-color", "#D9534F");
			$("#myUsername").css("color", "white");
		}
		else
		{
			$("#rivalBlock").css("background-color", "#D9534F");
			$("#rivalUsername").css("color", "white");
			$("#myBlock").css("background-color", "white");
			$("#myUsername").css("color", "#325D88");

		}
	}

	function updateBoard()
	{
		if (mines == 0)
		{
			clearInterval(myInterval);
			return;
		}

		$.ajax({
			url: "/game/data", 
			type: "POST",
			async: false
		}).done(function(resp)
		{	
			resp = JSON.parse(resp);
			//console.log(resp);
			
			r = resp.R;
			c = resp.C;
			turn = resp.Turn;
			score[0] = resp.Score[0];
			score[1] = resp.Score[1];
			mines = resp.Mines;
			username = resp.Username;
			players = resp.Players

			for (var i = 0; i < r; i++)
			{
				for (var j = 0; j < c; j++)
				{
					var img;

					if (resp.StateBoard[i][j] != -1)
					{
						if (resp.Board[i][j] == -1)
						{
							if (resp.StateBoard[i][j] == 0)
								img = "redflag";
							else
								img = "blueflag";
						}
						else 
							img = "" + resp.Board[i][j];
					}
					else
						img = "hidden";

					$("#"+i+"_"+j).attr("src", "../images/"+img+".png");
				}
			}

			updateHTML();

			if (mines == 0)
				$("#winner").html("Ganador: "+score[0]>score[1]?"Player 0":"Player 1");
		});
	}

	function initBoard()
	{
		$.ajax({
			url: "/game/data", 
			type: "POST",
			async: false
		}).done(function(resp)
		{	
			resp = JSON.parse(resp);
			//console.log(resp);
			
			r = resp.R;
			c = resp.C;
			turn = resp.Turn;
			score[0] = resp.Score[0];
			score[1] = resp.Score[1];
			mines = resp.Mines;
			username = resp.Username;
			players = resp.Players;
			myPos = (resp.Username == resp.Players[0] ? 0 : 1);
			rivalPos = (myPos == 0 ? 1 : 0);

			var mytable = "";
			mytable += "<table cellspacing='0' cellpadding='0'>"
			for (var i = 0; i < r; i++)
			{
				mytable += "<tr>";
				for (var j = 0; j < c; j++)
				{
					var img;

					if (resp.StateBoard[i][j] != -1)
					{
						if (resp.Board[i][j] == -1)
						{
							if (resp.StateBoard[i][j] == 0)
								img = "redflag";
							else
								img = "blueflag";
						}
						else 
							img = "" + resp.Board[i][j];
					}
					else
						img = "hidden";

					mytable += "<td> <img id='"+i+"_"+j+"' name='cell' src='../images/"+img+".png' style='width:25px;height:25px;display:block;'> </td>"
				}
				mytable += "</tr>";
			}
			mytable += "</table>";

			$("#board").html(mytable);
			$("#myUsername").html(username);
			updateHTML();
		});
	}

	$("[name='cell']").click(function() 
	{
		if (mines == 0)
			return;

		$.ajax({
			url: "/game/move",
			type: "POST",
			data: {move: $(this).attr("id")}
		}).done(function(resp) 
		{
			resp = JSON.parse(resp);
			
			if (resp != null)
			{
				var keep = false;
				for (var i = 0; i < resp.length; i++) 
				{
					var img;
					if (resp[i].Val == -1)
					{
						img = turn == 0 ? "redflag" : "blueflag";
						score[turn]++;
						mines--;
						keep = true;
					}
					else
						img = "" + resp[i].Val;
						
					$("#"+resp[i].X+"_"+resp[i].Y).attr("src", "../images/"+img+".png");
				}

				if (keep == false)
				{
					turn = (turn == 1 ? 0 : 1);
					$("#players").html(stringPlayers());
				}
				else
				{
					updateHTML();

					if (mines == 0)
						$("#winner").html("Ganador: "+score[0]>score[1]?"Player 0":"Player 1");
				}
			}
		});
	});
});
