$(document).ready(function () {
    pollQueueLastUpdated();

    // Constructing the suggestion engine
    var engine = new Bloodhound({
        datumTokenizer: Bloodhound.tokenizers.whitespace,
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        remote: {
            url: `/api/v1/spotify/search/playlist?session=${$('#currentSessionJoinId').val()}&limit=50&query=%%query%%`,
            wildcard: '%%query%%',
            transform: function (response) {
                return response.results;
            }
        }
    });

    // Initializing the typeahead
    $('#playlistSearchInput').typeahead({
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
                    return `<div class="clickable" onclick="changeFallbackPlaylist('${suggestionData.spotifyPlaylistId}')">
                                <div class="media">
                                    <img src="${suggestionData.imageThumbnailUrl}" class="mr-3" alt="...">
                                    <div class="media-body">
                                        <h5 class="mt-0">${suggestionData.name}</h5>
                                        <p>${suggestionData.createdBy} - ${suggestionData.trackCount} tracks</p>
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
});

function changeFallbackPlaylist(playlistId) {
    $('#changeFallbackPlaylistIdInput').val(playlistId);
    $('#changeFallbackPlaylistForm').submit();
}