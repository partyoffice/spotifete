let currentSessionJoinId;

$(document).ready(function () {
    currentSessionJoinId = $('#currentSessionJoinId').val();

    // Constructing the suggestion engine
    var engine = new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        remote: {
            url: `/api/v1/spotify/search/track?session=${currentSessionJoinId}&limit=5&query=%%query%%`,
            wildcard: '%%query%%',
            transform: function (response) {
                return response.results;
            }
        }
    });

    // Initializing the typeahead
    $('.typeahead').typeahead({
            hint: false,
            highlight: true,
            minLength: 3,
            classNames: {
                menu: 'card',
                dataset: 'list-group list-group-flush',
                suggestion: 'list-group-item',
                empty: ''
            }
        },
        {
            name: 'api-search',
            source: engine,
            limit: 10,
            display: function () {
                // Clear search input when selecting a suggestion
                return ''
            },
            templates: {
                suggestion: function (suggestionData) {
                    return `<div class="clickable" onclick="requestTrack('${suggestionData.trackId}')">
                                <div class="media">
                                    <img src="${suggestionData.albumImageThumbnailUrl}" class="mr-3" alt="...">
                                    <div class="media-body">
                                        <h5 class="mt-0">${suggestionData.trackName}</h5>
                                        <p>${suggestionData.artistName} - ${suggestionData.albumName}</p>
                                    </div>
                                </div>
                            </div>`;
                },
                pending: function () {
                    return '<div class="card-body">Searching...</div>';
                },
                notFound: function () {
                    return '<div class="card-body">No results :/</div>';
                },
                footer: function () {
                    return '<div class="card-footer">Search results via Spotify</div>'
                }
            }
        });
});

async function requestTrack(trackId) {
    const requestBody = JSON.stringify({
        trackId: trackId.toString()
    });

    const response = await fetch(`/api/v1/sessions/${currentSessionJoinId}/request`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: requestBody
    });

    if (response.status === 200) {
        console.log('request success. reloading page.');
        location.reload();
    } else {
        responseBody = await response.json();
        alert(`Could not add request: ${responseBody.message}`);
    }
}
