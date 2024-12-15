function toggleTagEditor() {	
	if (document.getElementById("tags_editor").hidden == true) {
		initializeTagEditor()
		document.getElementById("tags").hidden = true
		document.getElementById("tags_editor").hidden = false	
	} else {
		document.getElementById("tags_editor").hidden = true
		document.getElementById("tags").hidden = false
	}
}

function initializeTagEditor() {
	var tags = '';
	var tagList = document.getElementById("tags").getElementsByTagName("li");

	for (let i = 0; i < tagList.length; i++) {
		var tag = tagList[i].innerText;
		tags += tag + "\n"
	}
	document.getElementById("tags_edit_list").value = tags
}