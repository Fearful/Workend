'use strict';

describe('Doppler Editor', function() {
  browser.get('index.html');
  it('should automatically redirect to /index when location hash/fragment is empty', function() {
    expect(browser.getLocationAbsUrl()).toMatch("/");
  });
});
