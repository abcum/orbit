// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orbit

func init() {

	src := []byte("!function(e){function t(){}function n(e,t){return function(){e.apply(t,arguments)}}function o(e){if('object'!=typeof this)throw new TypeError('Promises must be constructed via new');if('function'!=typeof e)throw new TypeError('not a function');this._state=0,this._handled=!1,this._value=void 0,this._deferreds=[],s(e,this)}function i(e,t){for(;3===e._state;)e=e._value;return 0===e._state?void e._deferreds.push(t):(e._handled=!0,void a(function(){var n=1===e._state?t.onFulfilled:t.onRejected;if(null===n)return void(1===e._state?r:f)(t.promise,e._value);var o;try{o=n(e._value)}catch(i){return void f(t.promise,i)}r(t.promise,o)}))}function r(e,t){try{if(t===e)throw new TypeError('A promise cannot be resolved with itself.');if(t&&('object'==typeof t||'function'==typeof t)){var i=t.then;if(t instanceof o)return e._state=3,e._value=t,void u(e);if('function'==typeof i)return void s(n(i,t),e)}e._state=1,e._value=t,u(e)}catch(r){f(e,r)}}function f(e,t){e._state=2,e._value=t,u(e)}function u(e){2===e._state&&0===e._deferreds.length&&setTimeout(function(){e._handled||d(e._value)},1);for(var t=0,n=e._deferreds.length;n>t;t++)i(e,e._deferreds[t]);e._deferreds=null}function c(e,t,n){this.onFulfilled='function'==typeof e?e:null,this.onRejected='function'==typeof t?t:null,this.promise=n}function s(e,t){var n=!1;try{e(function(e){n||(n=!0,r(t,e))},function(e){n||(n=!0,f(t,e))})}catch(o){if(n)return;n=!0,f(t,o)}}var l=setTimeout,a='function'==typeof setImmediate&&setImmediate||function(e){l(e,1)},d=function(e){'undefined'!=typeof console&&console&&console.warn('Possible Unhandled Promise Rejection:',e)};o.prototype['catch']=function(e){return this.then(null,e)},o.prototype.then=function(e,n){var r=new o(t);return i(this,new c(e,n,r)),r},o.all=function(e){var t=Array.prototype.slice.call(e);return new o(function(e,n){function o(r,f){try{if(f&&('object'==typeof f||'function'==typeof f)){var u=f.then;if('function'==typeof u)return void u.call(f,function(e){o(r,e)},n)}t[r]=f,0===--i&&e(t)}catch(c){n(c)}}if(0===t.length)return e([]);for(var i=t.length,r=0;r<t.length;r++)o(r,t[r])})},o.resolve=function(e){return e&&'object'==typeof e&&e.constructor===o?e:new o(function(t){t(e)})},o.reject=function(e){return new o(function(t,n){n(e)})},o.race=function(e){return new o(function(t,n){for(var o=0,i=e.length;i>o;o++)e[o].then(t,n)})},o._setImmediateFn=function(e){a=e},o._setUnhandledRejectionFn=function(e){d=e},'undefined'!=typeof module&&module.exports?module.exports=o:e.Promise||(e.Promise=o)}(this);")

	Add("promise", src)

}
