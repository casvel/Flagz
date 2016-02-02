$('document').ready(function()
{
	/***** Variables *****/
	var r, c;
	var turn;
	var players = ["", ""];
	var hasRival = false;
	var score = [0, 0];
	var mines;
	var username;
	var myPos, rivalPos;
	var xhover = -10, yhover = -10;
	var bombActive = false, hasBomb = true, rivalHasBomb = true;
	var dX = [0, 1, 1, 1, 0, -1, -1, -1], dY = [1, 1, 0, -1, -1, -1, 0, 1];
	var myInterval = setInterval(updateBoard, 1000);
	var lastX = -1, lastY = -1;
	/***** END Variables *****/

	/***** "main" *****/
	initBoard();
	/***** END "main" *****/

	var myInterval = setInterval(updateBoard, 1000); 

	$(window).unload(function(){
		$.ajax({
			url: "/game/exit", 
			type: "POST",
			async: false
		});
	});	

	$('.panel-body').hide();
	$('<h5>'+ players[1] + '</h5>').appendTo('.panel-heading');

	$('.panel-heading').on('click', function()
	{
		$('.panel-body').toggle();
	});
	
	$(document).ready(function () { 
	     $('#sendMsg').click(function () {
	         var msg = $('textarea').val();		
	         $.ajax({
	           url: '/game/chat',
	           type: 'post',
	           dataType: 'html',
	           data : { ajax_post_data: msg},
	           success : function(data) {
	             alert(data);
	             $('.containerwell').html(data);
	           },
	         });
	      });
	});             

	/***** Auxiliar Functions *****/
	function stringPlayers()
	{
		if (turn == 0)
			return "<strong style='color:#D9534F'>"+players[0]+"</strong> vs "+players[1];
		else
			return players[0]+" vs <strong style='color:#D9534F'>"+players[1]+"</strong>";
	}

	function showEndModal()
	{
		if (score[myPos] > score[rivalPos])
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

	function updateVariables(resp)
	{
		// Esto tiene que ir antes de actualizar lastX y lastY.
		$("#"+lastX+"_"+lastY).css("border-style","none");

		r        = resp.R;
		c        = resp.C;
		turn     = resp.Turn;
		score[0] = resp.Score[0];
		score[1] = resp.Score[1];
		mines    = resp.Mines;
		username = resp.Username;
		players  = resp.Players
		
		lastX = resp.LastX;
		lastY = resp.LastY;

		hasBomb      = resp.HasBomb[myPos];
		rivalHasBomb = resp.HasBomb[rivalPos];

		hasRival = (players[0] != "" && players[1] != "");
	}

	function updateHTML()
	{
		$("#players").html(stringPlayers());
		$("#myScore").html("Score: "+score[myPos]);
		$("#rivalScore").html("Score: "+score[rivalPos]);
		$("#rivalUsername").html(players[rivalPos]);
		$("#mines").html("Mines: <strong>"+mines+"</strong>");
		$("#"+lastX+"_"+lastY).css({"border-style":"solid", "border-color":"#D9534F", "border-width":"2px"});
		if (!hasBomb)
		{
			$("#myBomb").addClass("disabled");
			$("#myBomb").removeClass("active");
			$("#imgMyBomb").attr("src", "../images/bomb-disabled.png");
		}
		if (!rivalHasBomb)
			$("#imgRivalBomb").attr("src", "../images/bomb-disabled.png");

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

		if (players[turn] == username && hasRival)
			return;

		if (mines == 0)
		{
			clearInterval(myInterval);
			return;
		}

		$.ajax({
			url: "/game/data", 
			type: "POST"
		}).done(function(resp)
		{	
			resp = JSON.parse(resp);
			//console.log(resp);

			updateVariables(resp);

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
					{
						if (bombActive)
						{
							var hoverBomb = (i == xhover && j == yhover);

							for (var k = 0; k < 8; ++k)
							{
								var x = xhover+dX[k];
								var y = yhover+dY[k];
						
								if (x < 0 || y < 0 || x == r || y == c)
									continue;
						
								if (x == i && y == j)
									hoverBomb = true;
							}

							if (hoverBomb)
								img = "hidden-hover";
							else
								img = "hidden";
						}
						else
						{
							if (xhover == i && yhover == j)
								img = "hidden-hover";
							else
								img = "hidden";
						}
					}

					$("#"+i+"_"+j).attr("src", "../images/"+img+".png");
				}
			}

			updateHTML();

			if (mines == 0)
				showEndModal();
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
			
			myPos = (resp.Username == resp.Players[0] ? 0 : 1);
			rivalPos = (myPos == 0 ? 1 : 0);
			updateVariables(resp);

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
	/***** END Auxiliar Functions *****/

	/***** Events *****/
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
		if (mines == 0 || players[turn] != username)
			return;


		var ids = $(this).attr("id");
		var id  = ids.split("_");
		var ok  = ($(this).attr("src") == "../images/hidden-hover.png");

		if (bombActive)
		{
			
			var xc = parseInt(id[0]);
			var yc = parseInt(id[1]);

			for (var k = 0; k < 8; ++k)
			{
				var x = xc + dX[k];
				var y = yc + dY[k];

				if (x < 0 || y < 0 || x == r || y == c)
					continue;

				ids += ","+x+"_"+y;
				if ($("#"+x+"_"+y).attr("src") == "../images/hidden-hover.png")
					ok = true;
			}
		}

		if (!ok)
			return;

		$.ajax({
			url: "/game/move",
			type: "POST",
			data: {move: ids, usedBomb: bombActive}
		}).done(function(resp) 
		{
			resp = JSON.parse(resp);
			
			if (resp != null)
			{
				$("#"+lastX+"_"+lastY).css("border-style","none");
				lastX = parseInt(id[0]);
				lastY = parseInt(id[1]);

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
					if (mines == 0)
						showEndModal();
				}

				if (bombActive)
				{	
					bombActive = false;
					hasBomb    = false;
				}
				updateHTML();
			}
		});
	});

	// Hover some cell
	$("[name='cell']").hover(
		function()
		{
			var id = $(this).attr("id").split("_");
			xhover = parseInt(id[0]);
			yhover = parseInt(id[1]);

			if ($(this).attr("src") == "../images/hidden.png")
				$(this).attr("src", "../images/hidden-hover.png");

			if (bombActive)
			{
				for (var i = 0; i < 8; ++i)
				{
					var x = xhover + dX[i];
					var y = yhover + dY[i];

					if (x < 0 || y < 0 || x == r || y == c)
						continue;

					if ($("#"+x+"_"+y).attr("src") == "../images/hidden.png")
						$("#"+x+"_"+y).attr("src", "../images/hidden-hover.png");
				}
			}
		},
		function()
		{
			if ($(this).attr("src") == "../images/hidden-hover.png")
				$(this).attr("src", "../images/hidden.png");

			if (bombActive)
			{
				for (var i = 0; i < 8; ++i)
				{
					var x = xhover + dX[i];
					var y = yhover + dY[i];

					if (x < 0 || y < 0 || x == r || y == c)
						continue;

					if ($("#"+x+"_"+y).attr("src") == "../images/hidden-hover.png")
						$("#"+x+"_"+y).attr("src", "../images/hidden.png");
				}
			}

			xhover = -10;
			yhover = -10;
		}
	);

	// Delete game when exit page
	$(window).unload(function(){
		$.ajax({
			url: "/game/exit", 
			type: "POST",
			async: false
		});
	});
	/***** END Events *****/
});
