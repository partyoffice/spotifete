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
            url: `/api/v2/session/id/${currentSessionJoinId}/search/track?limit=50&query=%%query%%`,
            wildcard: '%%query%%',
            transform: function (response) {
                return response.tracks;
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
                    return `<div class="clickable" onclick="requestTrack('${suggestionData.spotify_track_id}')">
                                <div class="media">
                                    <img src="${suggestionData.album_image_thumbnail_url}" class="mr-3" alt="${suggestionData.album_name}">
                                    <div class="media-body">
                                        <h5 class="mt-0">${suggestionData.track_name}</h5>
                                        <p>${suggestionData.artist_name} - ${suggestionData.album_name}</p>
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
        url: `/api/v2/session/id/${$('#currentSessionJoinId').val()}/queue/last-updated`
    }).done(function(data) {
        if (Date.parse(data.queue_last_updated) > Date.parse(queueLastUpdated)) {
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
