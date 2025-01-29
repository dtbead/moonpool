function toggleTagEditor() {	
	if (document.getElementById("tags_editor").hidden == true) {
		initializeTagEditor()
		document.getElementById("tags").hidden = true
		document.getElementById("tags_editor").hidden = false
		document.getElementById("tag_btn").innerText = "Cancel"
	} else {
		document.getElementById("tags_editor").hidden = true
		document.getElementById("tags").hidden = false
		document.getElementById("tag_btn").innerText = "Edit"
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


function replaceTags() {
	var urlOrigin = window.location.origin
	var archive_id = getArchiveID()
	if (archive_id == null) {
		setStatus("error: got invalid archive_id on tag edit")
		return
	}

	var tags = document.getElementById("tags_edit_list").value;

	const formData = new FormData();
	formData.append("id", archive_id)
	formData.append("tags", tags)

	fetch(urlOrigin + "/api/entry/"+archive_id+"/tags/replace", {
		method: 'POST',
		body: formData,
	})
	.then(response => {
		if (!response.ok) {
			setStatus("error: " + response.text)
			return
		}
		location.reload();
    })
}

// setStatus sets the status message in the bottom left corner.
function setStatus(msg) {
	document.getElementById("status").innerText = msg;
}

// getArchiveID gets the current archive_id of an entry according to the current URL.
function getArchiveID() {
	var url = window.location.href.split('?')[0];
	var archive_id = url.match(/\/(\d+)$/)[1];
	return archive_id
}