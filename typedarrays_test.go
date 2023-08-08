package goja

import (
	"testing"

	"github.com/dop251/goja/unistring"
)

func TestUint16ArrayObject(t *testing.T) {
	vm := New()
	buf := vm._newArrayBuffer(vm.global.ArrayBufferPrototype, nil)
	buf.data = make([]byte, 16)
	if nativeEndian == littleEndian {
		buf.data[2] = 0xFE
		buf.data[3] = 0xCA
	} else {
		buf.data[2] = 0xCA
		buf.data[3] = 0xFE
	}
	a := vm.newUint16ArrayObject(buf, 1, 1, nil)
	v := a.getIdx(valueInt(0), nil)
	if v != valueInt(0xCAFE) {
		t.Fatalf("v: %v", v)
	}
}

func TestArrayBufferGoWrapper(t *testing.T) {
	vm := New()
	data := []byte{0xAA, 0xBB}
	buf := vm.NewArrayBuffer(data)
	vm.Set("buf", buf)
	_, err := vm.RunString(`
	var a = new Uint8Array(buf);
	if (a.length !== 2 || a[0] !== 0xAA || a[1] !== 0xBB) {
		throw new Error(a);
	}
	`)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := vm.RunString(`
	var b = Uint8Array.of(0xCC, 0xDD);
	b.buffer;
	`)
	if err != nil {
		t.Fatal(err)
	}
	buf1 := ret.Export().(ArrayBuffer)
	data1 := buf1.Bytes()
	if len(data1) != 2 || data1[0] != 0xCC || data1[1] != 0xDD {
		t.Fatal(data1)
	}
	if buf1.Detached() {
		t.Fatal("buf1.Detached() returned true")
	}
	if !buf1.Detach() {
		t.Fatal("buf1.Detach() returned false")
	}
	if !buf1.Detached() {
		t.Fatal("buf1.Detached() returned false")
	}
	_, err = vm.RunString(`
	if (b[0] !== undefined) {
		throw new Error("b[0] !== undefined");
	}
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTypedArrayIdx(t *testing.T) {
	const SCRIPT = `
	var a = new Uint8Array(1);

	// 32-bit integer overflow, should not panic on 32-bit architectures
	if (a[4294967297] !== undefined) {
		throw new Error("4294967297");
	}

	// Canonical non-integer
	a["Infinity"] = 8;
	if (a["Infinity"] !== undefined) {
		throw new Error("Infinity");
	}
	a["NaN"] = 1;
	if (a["NaN"] !== undefined) {
		throw new Error("NaN");
	}

	// Non-canonical integer
	a["00"] = "00";
	if (a["00"] !== "00") {
		throw new Error("00");
	}

	// Non-canonical non-integer
	a["1e-3"] = "1e-3";
	if (a["1e-3"] !== "1e-3") {
		throw new Error("1e-3");
	}
	if (a["0.001"] !== undefined) {
		throw new Error("0.001");
	}

	// Negative zero
	a["-0"] = 88;
	if (a["-0"] !== undefined) {
		throw new Error("-0");
	}

	if (a[0] !== 0) {
		throw new Error("0");
	}

	a["9007199254740992"] = 1;
	if (a["9007199254740992"] !== undefined) {
		throw new Error("9007199254740992");
	}
	a["-9007199254740992"] = 1;
	if (a["-9007199254740992"] !== undefined) {
		throw new Error("-9007199254740992");
	}

	// Safe integer overflow, not canonical (Number("9007199254740993") === 9007199254740992)
	a["9007199254740993"] = 1;
	if (a["9007199254740993"] !== 1) {
		throw new Error("9007199254740993");
	}
	a["-9007199254740993"] = 1;
	if (a["-9007199254740993"] !== 1) {
		throw new Error("-9007199254740993");
	}

	// Safe integer overflow, canonical Number("9007199254740994") == 9007199254740994
	a["9007199254740994"] = 1;
	if (a["9007199254740994"] !== undefined) {
		throw new Error("9007199254740994");
	}
	a["-9007199254740994"] = 1;
	if (a["-9007199254740994"] !== undefined) {
		throw new Error("-9007199254740994");
	}
	`

	testScript(SCRIPT, _undefined, t)
}

func TestTypedArraySetDetachedBuffer(t *testing.T) {
	const SCRIPT = `
	let sample = new Uint8Array([42]);
	$DETACHBUFFER(sample.buffer);
	sample[0] = 1;

	assert.sameValue(sample[0], undefined, 'sample[0] = 1 is undefined');
	sample['1.1'] = 1;
	assert.sameValue(sample['1.1'], undefined, 'sample[\'1.1\'] = 1 is undefined');
	sample['-0'] = 1;
	assert.sameValue(sample['-0'], undefined, 'sample[\'-0\'] = 1 is undefined');
	sample['-1'] = 1;
	assert.sameValue(sample['-1'], undefined, 'sample[\'-1\'] = 1 is undefined');
	sample['1'] = 1;
	assert.sameValue(sample['1'], undefined, 'sample[\'1\'] = 1 is undefined');
	sample['2'] = 1;
	assert.sameValue(sample['2'], undefined, 'sample[\'2\'] = 1 is undefined');	
	`
	vm := New()
	vm.Set("$DETACHBUFFER", func(buf *ArrayBuffer) {
		buf.Detach()
	})
	vm.testScriptWithTestLib(SCRIPT, _undefined, t)
}

func TestTypedArrayDefinePropDetachedBuffer(t *testing.T) {
	const SCRIPT = `
	var desc = {
	  value: 0,
	  configurable: false,
	  enumerable: true,
	  writable: true
	};
	
	var obj = {
	  valueOf: function() {
		throw new Error("valueOf() was called");
	  }
	};
	let sample = new Uint8Array(42);
	$DETACHBUFFER(sample.buffer);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "0", desc),
	false,
	'Reflect.defineProperty(sample, "0", {value: 0, configurable: false, enumerable: true, writable: true} ) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "-1", desc),
	false,
	'Reflect.defineProperty(sample, "-1", {value: 0, configurable: false, enumerable: true, writable: true} ) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "1.1", desc),
	false,
	'Reflect.defineProperty(sample, "1.1", {value: 0, configurable: false, enumerable: true, writable: true} ) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "-0", desc),
	false,
	'Reflect.defineProperty(sample, "-0", {value: 0, configurable: false, enumerable: true, writable: true} ) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "2", {
	  configurable: true,
	  enumerable: true,
	  writable: true,
	  value: obj
	}),
	false,
	'Reflect.defineProperty(sample, "2", {configurable: true, enumerable: true, writable: true, value: obj}) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "3", {
	  configurable: false,
	  enumerable: false,
	  writable: true,
	  value: obj
	}),
	false,
	'Reflect.defineProperty(sample, "3", {configurable: false, enumerable: false, writable: true, value: obj}) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "4", {
	  writable: false,
	  configurable: false,
	  enumerable: true,
	  value: obj
	}),
	false,
	'Reflect.defineProperty("new TA(42)", "4", {writable: false, configurable: false, enumerable: true, value: obj}) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "42", desc),
	false,
	'Reflect.defineProperty(sample, "42", {value: 0, configurable: false, enumerable: true, writable: true} ) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "43", desc),
	false,
	'Reflect.defineProperty(sample, "43", {value: 0, configurable: false, enumerable: true, writable: true} ) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "5", {
	  get: function() {}
	}),
	false,
	'Reflect.defineProperty(sample, "5", {get: function() {}}) must return false'
	);
	
	assert.sameValue(
	Reflect.defineProperty(sample, "6", {
	  configurable: false,
	  enumerable: true,
	  writable: true
	}),
	false,
	'Reflect.defineProperty(sample, "6", {configurable: false, enumerable: true, writable: true}) must return false'
	);
	`
	vm := New()
	vm.Set("$DETACHBUFFER", func(buf *ArrayBuffer) {
		buf.Detach()
	})
	vm.testScriptWithTestLib(SCRIPT, _undefined, t)
}

func TestTypedArrayDefineProperty(t *testing.T) {
	const SCRIPT = `
	var a = new Uint8Array(1);

	assert.throws(TypeError, function() {
		Object.defineProperty(a, "1", {value: 1});
	});
	assert.sameValue(Reflect.defineProperty(a, "1", {value: 1}), false, "1");

	assert.throws(TypeError, function() {
		Object.defineProperty(a, "Infinity", {value: 8});
	});
	assert.sameValue(Reflect.defineProperty(a, "Infinity", {value: 8}), false, "Infinity");

	Object.defineProperty(a, "test", {value: "passed"});
	assert.sameValue(a.test, "passed", "string property");

	assert.throws(TypeError, function() {
		Object.defineProperty(a, "0", {value: 1, writable: false});
	}, "define non-writable");

	assert.throws(TypeError, function() {
		Object.defineProperty(a, "0", {get() { return 1; }});
	}, "define accessor");

	var sample = new Uint8Array([42, 42]);

	assert.sameValue(
	Reflect.defineProperty(sample, "0", {
	  value: 8,
	  configurable: true,
	  enumerable: true,
	  writable: true
	}),
	true
	);

	assert.sameValue(sample[0], 8, "property value was set");
	let descriptor0 = Object.getOwnPropertyDescriptor(sample, "0");
	assert.sameValue(descriptor0.value, 8);
	assert.sameValue(descriptor0.configurable, true, "configurable");
	assert.sameValue(descriptor0.enumerable, true);
	assert.sameValue(descriptor0.writable, true);
	`
	testScriptWithTestLib(SCRIPT, _undefined, t)
}

func TestTypedArrayGetInvalidIndex(t *testing.T) {
	const SCRIPT = `
	var TypedArray = Object.getPrototypeOf(Int8Array);
	var proto = TypedArray.prototype;
	Object.defineProperty(proto, "1", {
		get: function() {
			throw new Error("OrdinaryGet was called!");
		}
	});
	var a = new Uint8Array(1);
	assert.sameValue(a[1], undefined);
	assert.sameValue(a["1"], undefined);
	`
	testScriptWithTestLib(SCRIPT, _undefined, t)
}

func TestArrayBufferObjectMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         *arrayBufferObject
		expected    uint64
		newExpected uint64
		errExpected error
	}{
		{
			name:        "should have a value of SizeEmptyStruct given a nil array buffer object",
			val:         nil,
			expected:    SizeEmptyStruct,
			newExpected: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name:        "should have a value of SizeEmptyStruct given an empty array buffer object",
			val:         &arrayBufferObject{},
			expected:    SizeEmptyStruct,
			newExpected: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should account for baseObject overhead given an array buffer object with empty baseObject",
			val: &arrayBufferObject{
				baseObject: baseObject{},
			},
			// baseObject overhead
			expected: SizeEmptyStruct,
			// baseObject overhead
			newExpected: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should account for baseObject overhead and values given an array buffer object with non-empty baseObject",
			val: &arrayBufferObject{
				baseObject: baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
			},
			// baseObject overhead + key/value pair
			expected: SizeEmptyStruct + (4 + SizeInt),
			// baseObject overhead + key/value pair with string overhead
			newExpected: SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, 100, 100, 100, nil))
			if err != tc.errExpected {
				t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expected {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expected)
			}
			if newTotal != tc.newExpected {
				t.Fatalf("Unexpected new memory return. Actual: %v Expected: %v", newTotal, tc.newExpected)
			}
		})
	}
}

func TestTypedArrayObjectMemUsage(t *testing.T) {
	tests := []struct {
		name        string
		val         *typedArrayObject
		expected    uint64
		newExpected uint64
		errExpected error
	}{
		{
			name:        "should have a value of SizeEmptyStruct given a nil typed array object",
			val:         nil,
			expected:    SizeEmptyStruct,
			newExpected: SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should have a value of SizeEmptyStruct given an empty typed array object",
			val:  &typedArrayObject{},
			// typedArrayObject overhead + nil baseObject overhead
			expected: SizeEmptyStruct + SizeEmptyStruct,
			// typedArrayObject overhead + nil baseObject overhead
			newExpected: SizeEmptyStruct + SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should account for baseObject overhead given a typed array object with empty baseObject",
			val: &typedArrayObject{
				baseObject: baseObject{},
			},
			// typedArrayObject overhead + baseObject overhead
			expected: SizeEmptyStruct + SizeEmptyStruct,
			// typedArrayObject overhead + baseObject overhead
			newExpected: SizeEmptyStruct + SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should account for baseObject overhead and values given a typed array object with non-empty baseObject",
			val: &typedArrayObject{
				baseObject: baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
			},
			// typedArrayObject overhead + baseObject overhead + key/value pair
			expected: SizeEmptyStruct + SizeEmptyStruct + (4 + SizeInt),
			// typedArrayObject overhead + baseObject overhead + key/value pair with string overhead
			newExpected: SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected: nil,
		},
		{
			name: "should account for arrayBufferObject overhead and values given a typed array object with non-empty viewedArrayBuf",
			val: &typedArrayObject{
				viewedArrayBuf: &arrayBufferObject{
					baseObject: baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
				},
			},
			// typedArrayObject overhead + nil baseObject overhead + arrayBufferObject overhead + key/value pair
			expected: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeInt),
			// typedArrayObject overhead + nil baseObject overhead + arrayBufferObject overhead + key/value pair with string overhead
			newExpected: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected: nil,
		},
		{
			name: "should account for defaultCtor overhead given a typed array object with empty defaultCtor",
			val:  &typedArrayObject{defaultCtor: &Object{}},
			// typedArrayObject overhead + nil baseObject overhead + defaultCtor overhead
			expected: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct,
			// typedArrayObject overhead + nil baseObject overhead + defaultCtor overhead
			newExpected: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct,
			errExpected: nil,
		},
		{
			name: "should account for defaultCtor overhead and values given a typed array object with non-empty defaultCtor",
			val: &typedArrayObject{
				defaultCtor: &Object{
					self: &baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
				},
			},
			// typedArrayObject overhead + nil baseObject overhead + defaultCtor overhead + key/value pair
			expected: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeInt),
			// typedArrayObject overhead + nil baseObject overhead + defaultCtor overhead + key/value pair with string overhead
			newExpected: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, 100, 100, 100, nil))
			if err != tc.errExpected {
				t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expected {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expected)
			}
			if newTotal != tc.newExpected {
				t.Fatalf("Unexpected new memory return. Actual: %v Expected: %v", newTotal, tc.newExpected)
			}
		})
	}
}

func TestDataViewObjectMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *dataViewObject
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of SizeEmptyStruct given a nil data view object",
			val:            nil,
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should have a value of SizeEmptyStruct given an empty data view object",
			val:  &dataViewObject{},
			// typedArrayObject overhead + nil baseObject overhead
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
			// typedArrayObject overhead + nil baseObject overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for baseObject overhead given a data view object with empty baseObject",
			val: &dataViewObject{
				baseObject: baseObject{},
			},
			// typedArrayObject overhead + baseObject overhead
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
			// typedArrayObject overhead + baseObject overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for baseObject overhead and values given a data view object with non-empty baseObject",
			val: &dataViewObject{
				baseObject: baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
			},
			// typedArrayObject overhead + baseObject overhead + key/value pair
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + (4 + SizeInt),
			// typedArrayObject overhead + baseObject overhead + key/value pair with string overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected:    nil,
		},
		{
			name: "should account for arrayBufferObject overhead and values given a data view object with non-empty viewedArrayBuf",
			val: &dataViewObject{
				viewedArrayBuf: &arrayBufferObject{
					baseObject: baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
				},
			},
			// typedArrayObject overhead + nil baseObject overhead + arrayBufferObject overhead + key/value pair
			expectedMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeInt),
			// typedArrayObject overhead + nil baseObject overhead + arrayBufferObject overhead + key/value pair with string overhead
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, 100, 100, 100, nil))
			if err != tc.errExpected {
				t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expectedMem {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expectedMem)
			}
			if newTotal != tc.expectedNewMem {
				t.Fatalf("Unexpected new memory return. Actual: %v Expected: %v", newTotal, tc.expectedNewMem)
			}
		})
	}
}
