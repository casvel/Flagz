$(document).ready(function(){

	function getUrlParameter(sParam) 
	{
	    var sPageURL = decodeURIComponent(window.location.search.substring(1)),
	        sURLVariables = sPageURL.split('&'),
	        sParameterName,
	        i;

	    for (i = 0; i < sURLVariables.length; i++) {
	        sParameterName = sURLVariables[i].split('=');

	        if (sParameterName[0] === sParam) {
	            return sParameterName[1];
	        }
	    }

	    return undefined;
	}

	var param = getUrlParameter("success");
	if (param != undefined)
	{
		if (param == "0")
			$("#successRegister").alert("close");
		else
		{
			$("#failRegister").alert("close");
			$("#successRegister").fadeTo(2000, 500).slideUp(500, function(){
		    	$(this).alert("close");
			});
		}
	}
	else
	{
		$("#successRegister").alert("close");
		$("#failRegister").alert("close");
	}

	$("#linkLogin").click(function(){
		$("#sectionLogin").show();
		$("#sectionRegister").hide();
	});

	$("#linkRegister").click(function(){
		$("#sectionLogin").hide();
		$("#sectionRegister").show();
	});

});