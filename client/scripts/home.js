// Default map zoom.
var DEFAULT_ZOOM = 1;
// Default map center.
var DEFAULT_CENTER = new google.maps.LatLng(0, 0);
// Default StreetView zoom.
var DEFAULT_STREETVIEW_ZOOM = 1;
// Default camera distance in StreetView.
var STREET_VIEW_DISTANCE = 50;

var MAP_OPTIONS = {
  zoom: DEFAULT_ZOOM,
  center: DEFAULT_CENTER,
  mapTypeId: google.maps.MapTypeId.ROADMAP,
  navigationControl: true,
  mapTypeControl: true,
  streetViewControl: true,
};

var PANO_OPTIONS = {
  position: null,
  visible: true,
  addressControl: false,
  enableCloseButton: true,
  navigationControlOptions: {
    style: google.maps.NavigationControlStyle.ZOOM_PAN,
  },
};

function main() {
  var page = new Page();
  page.initialize();
}

function Page() {
}

Page.prototype.initialize = function() {
  // Get a handle on each control.
  this.mapCanvas = $("#map_canvas");
  this.createButton = $("#create_button");

  // Create Google Map.
  this.map = new google.maps.Map(this.mapCanvas[0], MAP_OPTIONS);

  // Add click handler for create button.
  this.createButton.click(bind(this, this.createButtonClick));
}

Page.prototype.showError = function(message) {
  alert(message);
}

Page.prototype.createButtonClick = function() {
  // Check that the map is in Street View mode.
  var sv = this.map.getStreetView();
  if (!sv.getVisible()) {
    this.showError("You must be currently viewing a Street View panorama!");
    return;
  }

  // Get the panorama ID.
  var panoId = sv.getPano();
  window.location.href = "create?panoid=" + panoId;
}

// Run main function once DOM is ready.
$(document).ready(main);
