$(document).ready(function(){

	
	var notsIds, newnots = 0;
	var notifInterval;

	UpdateNotifications();
	notifInterval = setInterval(UpdateNotifications, 500);

	function createNotification(id, type, info, seen)
	{
		var notif = "";

		if (type == "challenge")
		{
			var aux      = info.split(".");
			var rival    = aux[0];
			var gameId   = aux[1];
			var disabled = seen ? "disabled" : "";

			notif += "<li><a href='#dropdown1' data-toggle='tab'><font color='#29ABE0'><strong>"+rival+"</strong></font> has challenged you!</a></li>";
			notif += "<li id='notif-"+id+"' name='"+rival+"'>";
    		notif += "<div class='btn-group btn-group-justified' style='padding-left:10px;padding-right:10px'>";
	    	notif += "<a id='accept-challenge' name='"+gameId+"' class='btn btn-primary btn-xs'"+disabled+">";
	    	notif += "<span class='glyphicon glyphicon-ok' aria-hidden='true'></span> Accept";
	    	notif += "</a>";
	    	notif += "<a id='reject-challenge' name='"+gameId+"' class='btn btn-danger btn-xs'"+disabled+">";
	    	notif += "<span class='glyphicon glyphicon-remove' aria-hidden='true'></span> Reject";
	    	notif += "</a>";
	    	notif += "</div>";
	    	notif += "</li>";
	    	notif += "<li class='divider'></li>";
		}
		else if (type == "reject")
		{
			notif += "<li><a href='#dropdown1' data-toggle='tab'><font color='#29ABE0'><strong>"+info+"</strong></font> has rejected your challenge</a></li>";
		}

		return notif;
	}

	function UpdateNotifications()
	{
		$.ajax({
			url: "misc/notification/get",
			type: "POST",
			dataType: "json",
			success: function(resp)
			{
				//console.log(resp);
				notsIds = [];
				newnots = 0;

				if (resp != null)
				{
					var notifs = "";

					for (var i = resp.length-1; i >= 0; i--)
					{
						var type = resp[i].Not.Type;
						var info = resp[i].Not.Info;
						var id   = resp[i].IdNot;
						var seen = resp[i].Not.Seen;

						notsIds.push({id:id, type:type});

						if (!seen)
							newnots++;

						notifs += createNotification(id, type, info, seen);
					}

					$("#notifications .dropdown-menu").html(notifs);
				}

				if (newnots > 0)
					$("#tag-nots").html("Notifications <span class='badge' style='background-color:#6B9430'>"+newnots+"</span><span class='caret'></span>");
				else
					$("#tag-nots").html("Notifications <span class='caret'></span>");
			}
		});
	}

	$("body").on("click", "#accept-challenge:not([disabled])", function()
	{
		var notId  = $(this).closest("li").attr("id").split("-")[1];
		var gameId = $(this).attr("name");

		$.ajax({
			url: "misc/notification/seen",
			type: "POST",
			data: {notId:notId}
		});

		$.ajax({
			url: "/game/joinGame",
			type: "POST",
			data: {id:gameId}
		}).done(function(){
			window.location.href = "/game";
		});
	});


	$("body").on("click", "#reject-challenge:not([disabled])", function()
	{
		var notif  = $(this).closest("li");
		var notId  = notif.attr("id").split("-")[1];
		var rival  = notif.attr("name");
		var gameId = $(this).attr("name");

		$.ajax({
			url: "misc/notification/reject/game",
			type: "POST",
			data: {gameId:gameId, rival:rival}
		});

		$.ajax({
			url: "misc/notification/seen",
			type: "POST",
			data: {notId:notId}
		});
	});

	$("#notifications").click(function()
	{
		for (var i = 0; i < notsIds.length; i++)
			if (notsIds[i].type != "challenge")
				$.ajax({
					url: "misc/notification/seen",
					type: "POST",
					data: {notId:notsIds[i].id}
				});
	});
});