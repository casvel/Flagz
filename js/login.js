$(document).ready(function(){

	$("#linkLogin").click(function(){
		$("#sectionLogin").show();
		$("#sectionRegister").hide();
	});

	$("#linkRegister").click(function(){
		$("#sectionLogin").hide();
		$("#sectionRegister").show();
	});

});