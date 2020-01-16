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
    return $( '#joinIdInput' ).val().replace(/[\D]/g, '');
};

$( document ).ready(function() {
    if(navigator.userAgent.toLowerCase().indexOf('android') > -1) {
        if (window.confirm('Hey there!\n\nIt looks like you are using an android device.\n\nDo you want to install our App?')) {
            window.location.href='/app/android';
        }
    }

    $('#joinIdInput').on('input', function () {
        const joinIdConfirmButton = $('#joinIdConfirmButton');
        if (isJoinIdInputValid()) {
            joinIdConfirmButton.removeAttr('disabled');
            joinIdConfirmButton.removeAttr('title')
        } else {
            joinIdConfirmButton.attr('disabled', 'disabled');
            joinIdConfirmButton.attr('title', 'Enter a valid join id first.');
        }
    })
});