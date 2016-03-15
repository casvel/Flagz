$('document').ready(function()
{
	/***** Variables *****/
	var dX = [0, 1, 1, 1, 0, -1, -1, -1], dY = [1, 1, 0, -1, -1, -1, 0, 1];

	var game;	
	var username;
	var myPos, rivalPos;	
	
	var canUpdate = true;
	var hasRival = false;
	var bombActive = false;
	var xhover = -10, yhover = -10;

	var myInterval = setInterval(updateBoard, 1000);	
	/***** END Variables *****/

	/***** "main" *****/
	initBoard();
	$('.panel-body').hide();	
	/***** END "main" *****/


	/***** Functions *****/
	function stringPlayers()
	{
		if (game.Turn == 0)
			return "<strong style='color:#D9534F'>"+game.Players[0]+"</strong> vs "+game.Players[1];
		else
			return game.Players[0]+" vs <strong style='color: rgba(0, 78, 181,.9)'>"+game.Players[1]+"</strong>";
	}

	function gameOver()
	{
		return game.Score[0] >= 26 || game.Score[1] >= 26;
	}

	function endGame()
	{
		showMines();
		showEndModal();
		document.title = "Flagz - Game Over";	
	}

	function showEndModal()
	{
		if (game.Score[myPos] > game.Score[rivalPos])
		{
			$("#modal-end .modal-title").html("You Win!");
			$("#modal-end .modal-body").html("<p><h4 style='text-align:center'> You're awesome! <br><br> <img src='../images/win-baby.jpg' align='middle' style='width:60%;height:50%'> </h4> </p>");
		}
		else
		{
			$("#modal-end .modal-title").html("You Lose!");
			$("#modal-end .modal-body").html("<p><h4 style='text-align:center'> You suck! <br><br> <img src='../images/loser.jpg' align='middle' style='width:60%;height:50%'> </h4> </p>");
		}

		$("#modal-end").modal();
	}

	function showMines()
	{
		for (var i = 0; i < game.R; i++)
			for (var j = 0; j < game.C; j++)
				if (game.StateBoard[i][j] == -1 && game.Board[i][j] == -1)
					$("#"+i+"_"+j).attr("src", "../images/mine.png");
	}

	function updateVariables(resp)
	{
		game     = resp.Game;
		hasRival = (game.Players[rivalPos] != "");
	}

	function updateHTML()
	{
		$("#players").html(stringPlayers());
		$("#myScore").html("Score: "+game.Score[myPos]);
		$("#rivalScore").html("Score: "+game.Score[rivalPos]);
		$("#rivalUsername").html(game.Players[rivalPos]);
		$("#mines").html("Mines: <strong>"+game.MinesLeft+"</strong>");

		if (!game.HasBomb[myPos])
		{
			$("#myBomb").addClass("disabled");
			$("#myBomb").removeClass("active");
			$("#imgMyBomb").attr("src", "../images/bomb-disabled.png");
		}
		if (!game.HasBomb[rivalPos])
			$("#imgRivalBomb").attr("src", "../images/bomb-disabled.png");

		if (game.Players[game.Turn] == username)
		{
			if(game.Turn == 0)	
			{			
				$("#myBlock").css("background-color", "#D9534F");
				$("#rivalBlock").css("background-color", "white");
				$("#rivalUsername").css("color", "rgba(0, 78, 181,.9)");
				$("#myUsername").css("color", "white");
			}
			else
			{
				$("#myBlock").css("background-color", "rgba(0, 78, 181,.9)");
				$("#rivalBlock").css("background-color", "white");
				$("#rivalUsername").css("color", "#D9534F");
				$("#myUsername").css("color", "white");

			}
		}
		else
		{
			if(game.Turn == 0)
			{
				$("#myBlock").css("background-color", "white");
				$("#rivalBlock").css("background-color", "#D9534F");
				$("#rivalUsername").css("color", "white");
				$("#myUsername").css("color", "rgba(0, 78, 181,.9)");
			}
			else
			{
				$("#myBlock").css("background-color", "white");
				$("#rivalBlock").css("background-color", "rgba(0, 78, 181,.9)");
				$("#rivalUsername").css("color", "white");
				$("#myUsername").css("color", "#D9534F");
			}				
		}

		if (gameOver())
			document.title = "Flagz - Game Over";
		else
		{
			if (game.Players[game.Turn] == username)
				document.title = "Flagz - Your turn!";
			else
				document.title = "Flagz - Game";
		}
	}

	function getImg(i, j)
	{
		var img;
		var bgcolor;
		var bcolor;

		if (game.StateBoard[i][j] != -1)
		{
			if (game.Board[i][j] == -1)
			{
				if (game.StateBoard[i][j] == 0)				
					img = "redflag";									
				else			
					img = "blueflag";
				
				bgcolor = " #74C8EB";
				bcolor  = " 1px solid rgba(41,171,224,.9) ";
			}
			else 
			{
				img    = "" + game.Board[i][j];
				bcolor = " 1px solid rgba(132,132,132, .9) ";
				if(img == "0")
					bgcolor = " rgba(132,132,132, .8) ";
				else
					bgcolor = " rgba(132,132,132, .7) ";				
			}

			if (i == game.LastX && j == game.LastY)
			{
				bgcolor = " #E68F8C ";
				bcolor  = " 1px solid rgba(223,1,1, .7) ";
			}
		}
		else
		{
			img = "0";
			bgcolor = " rgba(0,0,255, .7) ";
			bcolor  = " 1px solid rgba(0,0,255, .9) ";

			if (xhover == i && yhover == j)
			{
				bgcolor = " rgba(0,0,255, .2) ";
				bcolor  = " 1px solid rgba(0,0,255, .9) ";
			}
			else if (bombActive)
			{
				var hover = false;
				for (var k = 0; k < 8; ++k)
				{
					var nx = xhover+dX[k], ny = yhover+dY[k];
					if (nx == i && ny == j)
					{
						bgcolor = " rgba(0,0,255, .2) ";
						bcolor  = " 1px solid rgba(0,0,255, .9) ";
					}
				}
			}
		}

		return [img, bgcolor, bcolor];
	}

	function updateBoard()
	{

		if (game.Players[game.Turn] == username || gameOver())
		{

			$.ajax({
				url: "/game/data",
				type: "POST",
				dataType: "json"
			}).done(function(resp)
			{
				game.Players = resp.Game.Players;
				updateHTML();
			});
		}
		else
		{
			$.ajax({
				url: "/game/data", 
				type: "POST",
				dataType: "json"
			}).done(function(resp)
			{	
				updateVariables(resp);

				for (var i = 0; i < game.R; i++)
					for (var j = 0; j < game.C; j++)
					{
						var img = getImg(i, j);
						$("#"+i+"_"+j).attr("src", "../images/"+img[0]+".png");		
						$("#"+i+"_"+j).css({'background-color' : img[1] , 'border' : img[2]});			
					}

				updateHTML();

				if (gameOver())
					endGame();
			});
		}
	}

	function initBoard()
	{
		$.ajax({
			url: "/game/data", 
			type: "POST",
			dataType: "json",
			async: false
		}).done(function(resp)
		{	
			username = resp.Username;
			myPos    = (username == resp.Game.Players[0] ? 0 : 1);
			rivalPos = (myPos == 0 ? 1 : 0);

			updateVariables(resp);

			var mytable = "";
			mytable += "<table cellspacing='0' cellpadding='0'>"
			for (var i = 0; i < game.R; i++)
			{
				mytable += "<tr>";
				for (var j = 0; j < game.C; j++)
				{
					var img = getImg(i, j);
					mytable += "<td> <img id='"+i+"_"+j+"' name='cell' src='../images/"+img[0]+".png' style='width:25px;height:25px;display:block; background-color:"+  img[1] + "; border: " + img[2] + ";'> </td>";					
				}
				mytable += "</tr>";
			}
			mytable += "</table>";

			$("#board").html(mytable);
			$("#myUsername").html(username);
			updateHTML();
		});
	}

	function bfs(start)
	{
		var queue = Array(game.R*game.C);
		var h = 0; t = -1;

		for (var i = 0; i < start.length; i++)
		{
			t++;
			queue[t] = start[i];
			game.StateBoard[start[i][0]][start[i][1]] = game.Turn;
		}

		var visited = [];

		while (t >= h)
		{
			var x = queue[h][0];
			var y = queue[h][1];
			h++;

			if (game.Board[x][y] == -1)
			{
				game.Score[game.Turn]++;
				game.MinesLeft--;
				keep = true;
			}

			visited.push(x, y);
			var img = getImg(x, y);
			$("#"+x+"_"+y).attr("src", "../images/"+img[0]+".png");
			$("#"+x+"_"+y).css({'background-color' : img[1] , 'border' : img[2]});

			if (game.Board[x][y] == 0)
				for (var k = 0; k < 8; k++)
				{
					var nx = x+dX[k], ny = y+dY[k];
					if (nx < 0 || nx == game.R || ny < 0 || ny == game.C || game.StateBoard[nx][ny] != -1)
						continue;

					game.StateBoard[nx][ny] = game.Turn;
					t++;
					queue[t] = [nx, ny];
				}
		}

		return visited;
	}
	/***** END Functions *****/	

	/***** Events *****/

	// Show chat
	$('.panel-heading').on('click', function()
	{
		$('.panel-body').toggle();		
	});    

	// Activate bomb
	$("#myBomb").click(function(){

		if ($(this).hasClass("active"))
		{
			bombActive = false;
			$(this).trigger("blur");
			$(this).removeClass("active");
		}
		else
		{
			bombActive = true;
			$(this).trigger("blur");
			$(this).addClass("active");
		}
	});

	// Click on some cell
	$("[name='cell']").click(function() 
	{
		if (gameOver() || game.Players[game.Turn] != username)
			return;

		var ids = [];
		var ok = false;
		
		var id  = $(this).attr("id").split("_");
		id[0] = parseInt(id[0]);
		id[1] = parseInt(id[1]);

		if (game.StateBoard[id[0]][id[1]] == -1)
		{
			ids.push([id[0], id[1]]);
			ok = true;
		}

		if (bombActive)
		{
			var xc = id[0];
			var yc = id[1];

			for (var k = 0; k < 8; ++k)
			{
				var x = xc + dX[k];
				var y = yc + dY[k];

				if (x < 0 || y < 0 || x == game.R || y == game.C)
					continue;

				if (game.StateBoard[x][y] == -1)
				{
					ids.push([x, y]);
					ok = true;
				}
			}
		}

		if (!ok)
			return;

		canUpdate = false;

		var lastX = game.LastX, lastY = game.LastY;
		game.LastX = id[0];
		game.LastY = id[1];
		if (lastX != -1)
		{
			var img = getImg(lastX, lastY);
			$("#"+lastX+"_"+lastY).attr("src", "../images/"+img[0]+".png");
			$("#"+lastX+"_"+lastY).css({'background-color':img[1], 'border':img[2]});
		}

		keep = false;
		visited = bfs(ids);

		$.ajax({
			url: "/game/move",
			type: "POST",
			data: {visited:visited, usedBomb:bombActive, lastX:game.LastX, lastY:game.LastY}
		});

		if (keep == false)
		{
			game.Turn = (game.Turn == 1 ? 0 : 1);
			$("#players").html(stringPlayers());
		}

		if (bombActive)
		{	
			bombActive = false;
			game.HasBomb[myPos] = false;
		}

		if (gameOver())
			endGame();

		updateHTML();

		canUpdate = true;
	});

	// Hover some cell
	$("[name='cell']").hover(
		function()
		{
			var id = $(this).attr("id").split("_");
			xhover = parseInt(id[0]);
			yhover = parseInt(id[1]);

			if (game.StateBoard[xhover][yhover] == -1)
			{
				$("#"+xhover+"_"+yhover).css({'background-color' : 'rgba(0,0,255, .2)'});
			}

			if (bombActive)
			{
				for (var i = 0; i < 8; ++i)
				{
					var x = xhover + dX[i];
					var y = yhover + dY[i];

					if (x < 0 || y < 0 || x == game.R || y == game.C)
						continue;

					if (game.StateBoard[x][y] == -1)
					{
						$("#"+x+"_"+y).css({'background-color':'rgba(0,0,255, .2)'});
					}
				}
			}
		},
		function()
		{
			var id = $(this).attr("id").split("_");
			x = parseInt(id[0]);
			y = parseInt(id[1]);
			if (game.StateBoard[x][y] == -1)
			{
				$("#"+x+"_"+y).css({'background-color':'rgba(0,0,255, .7)'});
			}

			if (bombActive)
			{
				for (var i = 0; i < 8; ++i)
				{
					var x = xhover + dX[i];
					var y = yhover + dY[i];

					if (x < 0 || y < 0 || x == game.R || y == game.C)
						continue;

					if (game.StateBoard[x][y] == -1)
					{
						$("#"+x+"_"+y).css({'background-color':'rgba(0,0,255, .7)'});
					}
				}
			}

			xhover = -10;
			yhover = -10;
		}
	);

	// Delete game when exit page
	/*$(window).unload(function(){
		$.ajax({
			url: "/game/exit", 
			type: "POST",
			async: false
		});
	});*/
	/***** END Events *****/
});
