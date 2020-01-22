if(navigator.userAgent.toLowerCase().indexOf('android') > -1 && document.cookie.indexOf('appDeclined=true') === -1) {
    if (window.confirm('Hey there!\n\nIt looks like you are using an android device.\n\nDo you want to install our App?')) {
        window.location.href='/app/android';
    } else {
        document.cookie = "appDeclined=true; path=/;";
    }
}
