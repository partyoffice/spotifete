let currentSessionJoinId;
let queueLastUpdated;

$(document).ready(function () {
    currentSessionJoinId = $('#currentSessionJoinId').val();
    queueLastUpdated = $('#queueLastUpdated').val();

    pollQueueLastUpdated();

    // Constructing the suggestion engine
    var engine = new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        remote: {
            url: `/api/v1/spotify/search/track?session=${currentSessionJoinId}&limit=50&query=%%query%%`,
            wildcard: '%%query%%',
            transform: function (response) {
                return response.results;
            }
        }
    });

    const trackSearchInput = $('#trackSearchInput');
    // Initializing the typeahead
    trackSearchInput.typeahead({
            hint: false,
            highlight: true,
            minLength: 2,
            classNames: {
                menu: 'card text-left',
                dataset: 'list-group list-group-flush',
                suggestion: 'list-group-item',
                empty: ''
            }
        },
        {
            name: 'api-search',
            source: engine,
            limit: 100,
            display: function () {
                // Clear search input when selecting a suggestion
                return ''
            },
            templates: {
                suggestion: function (suggestionData) {
                    return `<div class="clickable" onclick="requestTrack('${suggestionData.SpotifyTrackId}')">
                                <div class="media">
                                    <img src="${suggestionData.AlbumImageThumbnailUrl}" class="mr-3" alt="...">
                                    <div class="media-body">
                                        <h5 class="mt-0">${suggestionData.TrackName}</h5>
                                        <p>${suggestionData.ArtistName} - ${suggestionData.AlbumName}</p>
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
                    return '<div class="card-footer">Search results via Spotify <span class="fab fa-spotify"></span></div>'
                }
            }
        });

    // After initializing, focus the input again
    trackSearchInput.focus();
});

function requestTrack(trackId) {
    $('#requestTrackIdInput').val(trackId);
    $('#submitRequestForm').submit();
}

function pollQueueLastUpdated() {
    $.ajax({
        url: `/api/v1/sessions/${currentSessionJoinId}/queuelastupdated`
    }).done(function(data) {
        if (Date.parse(data.queueLastUpdated) > Date.parse(queueLastUpdated)) {
            location.reload();
        } else {
            setTimeout(function(){
                pollQueueLastUpdated();
            }, 2000);
        }
    }).fail(function () {
        setTimeout(function(){
            pollQueueLastUpdated();
        }, 2000);
    });
}
