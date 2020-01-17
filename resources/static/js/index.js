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

function disableJoinIdConfirmButtonIfNeccessary() {
    const joinIdConfirmButton = $('#joinIdConfirmButton');
    if (isJoinIdInputValid()) {
        joinIdConfirmButton.removeAttr('disabled');
        joinIdConfirmButton.removeAttr('title')
    } else {
        joinIdConfirmButton.attr('disabled', 'disabled');
        joinIdConfirmButton.attr('title', 'Enter a valid join id first.');
    }
}

$( document ).ready(function() {
    $('#joinIdInput').on('input', function () {
        disableJoinIdConfirmButtonIfNeccessary();
    });
    disableJoinIdConfirmButtonIfNeccessary();
});