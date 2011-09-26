function bind(scope, fn) {
  var args = Array.prototype.slice.call(arguments, 2);
  return function() {
    return fn.apply(scope, args.concat(toArray(arguments)));
  }
}

function toArray(obj) {
  return Array.prototype.slice.call(obj);
}
