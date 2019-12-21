joinClicked = () => {
    const joinId = getJoinIdInputValue();
    if (isJoinIdInputValid()) {
        window.location.href = `session/view/${joinId}`

    } else {
        // TODO: do something nicer here
        alert('Invalid join id');
    }
};

isJoinIdInputValid = () => {
    return getJoinIdInputValue().match(/^\d{8}$/);
};

getJoinIdInputValue = () => {
    // TODO: Refactor this after splitting into 8 separate inputs
    return $( '#joinIdInput' ).val();
};

$( document ).ready(function() {
    if(navigator.userAgent.toLowerCase().indexOf('android') > -1) {
        if (window.confirm('Hey there!\n\nIt looks like you are using an android device.\n\nDo you want to install our App?')) {
            window.location.href=$('#appUrl').val();
        }
    }
});