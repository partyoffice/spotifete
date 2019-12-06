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
    const joinIdInputValue = getJoinIdInputValue();
    return joinIdInputValue.match(/^\d{8}$/);
};

getJoinIdInputValue = () => {
    // TODO: Refactor this after splitting into 8 separate inputs
    return $( '#joinIdInput' ).val();
};
