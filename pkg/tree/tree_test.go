/*
Copyright 2019 vChain, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tree

import (
	"crypto/sha256"
	"math"
	"math/bits"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// todo(leogr): double check
var testRoots = [][sha256.Size]byte{
	{219, 52, 38, 232, 120, 6, 141, 40, 210, 105, 182, 200, 113, 114, 50, 44, 229, 55, 43, 101, 117, 109, 7, 137, 0, 29, 52, 131, 95, 96, 28, 3},
	{203, 0, 152, 157, 148, 165, 105, 192, 166, 120, 174, 4, 43, 99, 220, 212, 98, 93, 185, 100, 64, 81, 127, 55, 166, 235, 121, 118, 234, 36, 237, 75},
	{16, 138, 42, 27, 152, 29, 188, 188, 0, 92, 160, 102, 69, 31, 45, 208, 253, 66, 45, 12, 27, 228, 86, 89, 28, 106, 232, 98, 230, 129, 171, 0},
	{159, 74, 63, 194, 13, 65, 98, 220, 55, 212, 226, 61, 144, 120, 72, 115, 26, 118, 4, 63, 255, 246, 214, 146, 136, 191, 26, 191, 188, 255, 71, 142},
	{194, 61, 131, 90, 104, 212, 12, 11, 102, 40, 241, 221, 193, 150, 147, 161, 40, 205, 74, 194, 154, 0, 28, 225, 220, 95, 124, 236, 199, 93, 201, 93},
	{8, 239, 108, 232, 144, 63, 142, 120, 115, 93, 1, 78, 212, 227, 71, 208, 174, 214, 114, 48, 130, 114, 69, 69, 100, 62, 41, 146, 185, 79, 144, 22},
	{212, 106, 160, 209, 240, 129, 233, 20, 177, 199, 211, 44, 190, 79, 79, 142, 85, 107, 163, 199, 154, 52, 18, 221, 71, 231, 188, 194, 215, 121, 40, 54},
	{59, 133, 169, 98, 108, 28, 203, 100, 198, 185, 94, 199, 250, 100, 136, 141, 239, 226, 207, 18, 227, 158, 119, 225, 8, 18, 206, 95, 203, 156, 181, 142},
	{137, 149, 148, 123, 40, 240, 236, 154, 207, 109, 73, 253, 255, 127, 22, 100, 221, 91, 129, 24, 86, 57, 14, 243, 151, 251, 55, 237, 184, 241, 28, 35},
	{20, 121, 163, 20, 251, 248, 22, 241, 234, 68, 135, 10, 119, 124, 148, 216, 172, 240, 217, 122, 64, 172, 79, 186, 125, 132, 103, 106, 218, 59, 95, 220},
	{159, 68, 64, 62, 223, 130, 163, 10, 62, 210, 44, 215, 104, 219, 101, 248, 170, 204, 252, 7, 253, 233, 0, 177, 92, 89, 222, 39, 89, 9, 25, 154},
	{244, 64, 208, 83, 76, 86, 93, 122, 125, 213, 84, 55, 249, 56, 219, 143, 160, 65, 194, 155, 149, 48, 68, 121, 191, 237, 172, 119, 173, 203, 47, 221},
	{230, 241, 183, 4, 147, 229, 54, 108, 188, 26, 67, 100, 244, 60, 178, 185, 105, 118, 140, 194, 27, 132, 70, 83, 225, 64, 111, 104, 217, 25, 98, 246},
	{54, 62, 115, 9, 147, 113, 180, 32, 23, 149, 250, 106, 208, 126, 215, 251, 240, 29, 203, 73, 142, 46, 253, 221, 94, 94, 28, 57, 145, 170, 240, 127},
	{64, 23, 242, 12, 165, 238, 128, 111, 164, 30, 29, 217, 11, 78, 243, 245, 133, 98, 122, 78, 3, 125, 91, 42, 26, 189, 41, 165, 246, 50, 38, 237},
	{172, 252, 161, 145, 220, 202, 153, 31, 84, 192, 124, 84, 112, 11, 40, 172, 134, 37, 231, 70, 20, 178, 218, 219, 150, 231, 174, 148, 105, 75, 237, 200},
	{27, 214, 19, 150, 206, 76, 2, 115, 11, 91, 199, 83, 80, 114, 171, 158, 226, 118, 22, 179, 57, 203, 249, 234, 236, 222, 241, 252, 100, 32, 3, 50},
	{217, 69, 166, 221, 133, 137, 88, 7, 179, 34, 117, 137, 55, 196, 187, 159, 24, 28, 140, 79, 212, 130, 141, 8, 208, 182, 117, 85, 182, 235, 94, 182},
	{86, 175, 231, 211, 212, 86, 196, 119, 152, 67, 109, 69, 9, 153, 148, 254, 209, 56, 116, 25, 233, 87, 8, 170, 165, 41, 211, 73, 240, 65, 78, 192},
	{93, 134, 112, 92, 208, 114, 170, 180, 91, 249, 138, 31, 76, 168, 54, 158, 80, 89, 152, 90, 182, 3, 102, 87, 142, 92, 91, 192, 251, 228, 206, 134},
	{148, 162, 121, 91, 174, 253, 13, 110, 165, 232, 129, 136, 99, 55, 28, 70, 133, 140, 138, 176, 78, 175, 114, 104, 150, 196, 67, 100, 97, 178, 14, 110},
	{254, 242, 193, 23, 98, 134, 121, 90, 189, 65, 120, 183, 28, 235, 91, 250, 171, 143, 37, 228, 78, 197, 150, 168, 194, 128, 134, 15, 254, 210, 88, 244},
	{149, 186, 141, 204, 219, 93, 0, 188, 36, 193, 53, 196, 84, 193, 46, 193, 52, 11, 227, 34, 171, 194, 101, 57, 4, 239, 130, 183, 190, 15, 230, 215},
	{89, 210, 31, 70, 78, 40, 185, 124, 237, 160, 225, 33, 139, 97, 155, 36, 55, 221, 190, 80, 129, 128, 190, 220, 30, 192, 208, 149, 126, 182, 146, 38},
	{156, 177, 180, 150, 216, 225, 43, 168, 230, 68, 69, 173, 152, 176, 203, 141, 10, 41, 125, 47, 58, 69, 53, 74, 2, 217, 144, 190, 99, 133, 62, 25},
	{146, 169, 229, 24, 254, 78, 120, 137, 42, 101, 234, 254, 31, 235, 146, 128, 215, 138, 153, 184, 112, 133, 204, 55, 167, 136, 89, 255, 95, 185, 109, 189},
	{202, 149, 99, 129, 185, 164, 239, 176, 72, 85, 205, 222, 51, 145, 113, 127, 124, 74, 214, 57, 228, 29, 67, 35, 60, 4, 157, 127, 243, 155, 231, 22},
	{21, 55, 6, 239, 39, 200, 0, 20, 127, 214, 73, 88, 134, 191, 7, 102, 6, 5, 168, 194, 250, 85, 22, 44, 21, 70, 73, 207, 209, 126, 194, 145},
	{57, 19, 70, 167, 204, 62, 0, 102, 122, 169, 212, 72, 99, 210, 206, 84, 203, 221, 102, 206, 212, 48, 107, 167, 26, 3, 112, 80, 150, 46, 12, 98},
	{88, 20, 84, 44, 114, 233, 189, 74, 72, 24, 216, 156, 246, 151, 119, 118, 154, 209, 237, 218, 184, 208, 249, 0, 160, 252, 61, 240, 15, 213, 28, 208},
	{142, 55, 243, 70, 57, 104, 216, 242, 214, 105, 5, 55, 125, 178, 22, 163, 105, 97, 188, 87, 73, 8, 5, 172, 67, 36, 182, 137, 55, 235, 225, 118},
	{4, 87, 113, 238, 55, 75, 79, 100, 192, 111, 211, 201, 5, 248, 103, 108, 5, 148, 215, 47, 6, 39, 23, 223, 184, 59, 47, 32, 198, 71, 114, 204},
	{183, 51, 62, 86, 124, 28, 19, 180, 26, 89, 154, 80, 187, 28, 244, 184, 149, 9, 240, 243, 184, 81, 135, 189, 141, 11, 153, 54, 100, 110, 93, 89},
	{121, 21, 167, 140, 79, 155, 81, 10, 165, 70, 172, 123, 5, 13, 24, 46, 25, 94, 197, 170, 184, 193, 222, 87, 79, 185, 36, 18, 18, 58, 33, 113},
	{58, 232, 115, 99, 111, 42, 183, 9, 17, 191, 38, 148, 186, 165, 76, 144, 164, 161, 124, 100, 220, 160, 136, 100, 208, 20, 68, 52, 211, 219, 162, 10},
	{200, 171, 181, 104, 4, 108, 50, 29, 106, 178, 150, 20, 110, 120, 229, 254, 111, 118, 20, 60, 40, 27, 113, 192, 18, 68, 241, 218, 251, 46, 164, 170},
	{230, 103, 98, 160, 12, 92, 26, 36, 9, 1, 91, 236, 22, 144, 154, 151, 173, 210, 231, 143, 250, 16, 6, 227, 194, 222, 128, 50, 8, 241, 59, 210},
	{202, 5, 18, 153, 224, 237, 6, 220, 140, 190, 231, 148, 112, 116, 17, 201, 248, 28, 207, 185, 62, 35, 120, 130, 165, 52, 29, 85, 170, 103, 46, 178},
	{79, 206, 226, 223, 158, 43, 157, 216, 176, 43, 144, 111, 118, 251, 175, 29, 184, 55, 248, 222, 214, 40, 123, 59, 104, 223, 157, 92, 39, 164, 123, 179},
	{103, 241, 208, 193, 52, 115, 32, 2, 187, 90, 202, 241, 105, 5, 193, 122, 178, 121, 92, 159, 235, 73, 39, 112, 209, 195, 103, 126, 131, 129, 48, 17},
	{7, 44, 152, 207, 249, 184, 8, 140, 120, 231, 14, 246, 121, 249, 173, 70, 151, 101, 154, 103, 52, 189, 62, 37, 72, 105, 143, 58, 0, 146, 14, 69},
	{240, 9, 22, 16, 211, 69, 141, 95, 79, 24, 87, 126, 114, 88, 166, 118, 52, 99, 55, 177, 104, 80, 71, 54, 155, 200, 253, 233, 239, 198, 226, 112},
	{68, 78, 104, 138, 127, 135, 173, 125, 68, 88, 27, 226, 104, 203, 154, 63, 196, 60, 250, 176, 62, 92, 151, 251, 185, 157, 43, 254, 104, 70, 149, 197},
	{243, 246, 41, 65, 36, 178, 247, 32, 105, 254, 83, 234, 77, 95, 231, 124, 25, 74, 26, 251, 30, 233, 204, 94, 2, 153, 186, 123, 28, 108, 156, 241},
	{101, 233, 136, 88, 37, 119, 203, 137, 55, 65, 103, 174, 88, 123, 164, 121, 11, 214, 118, 191, 140, 45, 198, 251, 166, 150, 159, 248, 67, 38, 110, 136},
	{223, 238, 53, 182, 163, 107, 107, 151, 48, 157, 43, 162, 131, 231, 37, 251, 37, 227, 25, 212, 223, 177, 199, 174, 9, 35, 5, 186, 129, 203, 42, 121},
	{57, 77, 87, 142, 201, 34, 44, 186, 74, 212, 81, 190, 249, 77, 181, 132, 145, 153, 64, 6, 24, 82, 129, 245, 39, 184, 29, 142, 172, 127, 18, 141},
	{209, 138, 128, 26, 125, 156, 83, 28, 159, 42, 222, 32, 136, 254, 53, 190, 21, 73, 172, 12, 39, 43, 20, 151, 234, 229, 132, 216, 152, 120, 98, 251},
	{55, 149, 166, 65, 36, 218, 65, 102, 120, 185, 218, 43, 136, 97, 146, 86, 60, 241, 55, 183, 232, 118, 75, 9, 64, 169, 157, 61, 234, 10, 117, 22},
	{27, 32, 187, 181, 183, 15, 5, 254, 197, 55, 96, 126, 241, 171, 37, 250, 226, 146, 163, 130, 150, 165, 126, 235, 45, 73, 111, 152, 225, 201, 110, 198},
	{145, 132, 27, 160, 78, 255, 112, 18, 95, 76, 250, 76, 142, 49, 149, 228, 77, 19, 84, 92, 43, 104, 172, 176, 65, 215, 224, 137, 12, 184, 11, 133},
	{47, 197, 89, 98, 18, 5, 161, 199, 17, 53, 43, 180, 69, 193, 250, 180, 60, 61, 145, 88, 250, 3, 84, 95, 134, 241, 10, 128, 109, 61, 100, 57},
	{199, 248, 187, 137, 69, 150, 131, 21, 219, 250, 110, 14, 240, 8, 235, 98, 191, 72, 153, 89, 224, 187, 116, 61, 128, 79, 68, 61, 50, 35, 116, 186},
	{48, 14, 93, 152, 3, 8, 65, 48, 59, 68, 189, 56, 87, 4, 46, 194, 23, 132, 233, 148, 64, 96, 143, 29, 124, 98, 240, 174, 178, 17, 19, 173},
	{226, 1, 90, 18, 56, 40, 234, 108, 81, 196, 15, 107, 147, 253, 248, 176, 143, 253, 109, 67, 94, 88, 174, 74, 32, 225, 226, 206, 165, 53, 25, 57},
	{154, 254, 114, 164, 219, 137, 201, 140, 116, 197, 92, 31, 152, 177, 212, 69, 158, 39, 115, 166, 224, 247, 8, 108, 89, 56, 222, 60, 239, 34, 136, 1},
	{208, 193, 67, 45, 66, 165, 202, 251, 79, 199, 109, 205, 111, 179, 72, 151, 140, 32, 197, 17, 3, 248, 200, 102, 208, 200, 32, 73, 85, 220, 191, 34},
	{65, 80, 3, 10, 123, 197, 68, 242, 13, 223, 148, 246, 236, 152, 44, 54, 19, 226, 70, 53, 173, 129, 115, 197, 158, 157, 35, 252, 4, 202, 250, 177},
	{40, 47, 7, 92, 220, 175, 17, 144, 248, 62, 69, 168, 162, 66, 85, 126, 175, 84, 161, 1, 57, 129, 41, 85, 71, 8, 142, 162, 117, 194, 112, 28},
	{218, 137, 202, 153, 33, 180, 39, 102, 94, 143, 47, 181, 93, 106, 144, 187, 155, 54, 209, 248, 235, 178, 180, 225, 167, 213, 239, 1, 197, 34, 69, 130},
	{230, 1, 28, 29, 57, 253, 109, 204, 242, 132, 102, 174, 147, 242, 55, 66, 16, 225, 148, 85, 95, 12, 242, 18, 122, 66, 219, 38, 254, 96, 96, 83},
	{81, 104, 25, 165, 128, 31, 222, 196, 41, 57, 197, 161, 135, 35, 171, 112, 254, 135, 221, 55, 201, 189, 39, 111, 219, 83, 168, 158, 110, 30, 226, 107},
	{179, 0, 143, 223, 208, 97, 70, 8, 249, 138, 223, 129, 17, 47, 168, 37, 235, 191, 136, 220, 150, 152, 153, 92, 237, 146, 103, 127, 214, 97, 56, 127},
	{57, 26, 49, 238, 67, 215, 145, 187, 225, 25, 50, 165, 162, 36, 39, 25, 193, 136, 116, 232, 28, 243, 47, 13, 141, 61, 123, 243, 27, 252, 166, 184},
	{201, 28, 32, 28, 254, 198, 253, 1, 154, 47, 153, 197, 239, 70, 111, 170, 110, 101, 116, 67, 4, 89, 193, 198, 193, 167, 171, 118, 222, 198, 83, 244},
}

func TestTree(t *testing.T) {
	tr := New(NewMemStore())

	for n := 0; n <= 64; n++ {
		err := tr.Add([]byte(strconv.FormatUint(uint64(n), 10)))
		assert.NoError(t, err)

		assert.Equal(t, n, tr.N())
		d := int(math.Ceil(math.Log2(float64(n + 1))))
		assert.Equal(t, d, tr.Depth())

		assert.Equal(t, testRoots[n], tr.Root())
	}
}

func BenchmarkLog2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := i
		_ = int(math.Ceil(math.Log2(float64(n))))
	}
}

func BenchmarkLog2bits(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := uint64(i)
		_ = bits.Len64(n - 1)
	}
}

func BenchmarkTreeAdd(b *testing.B) {
	tr := New(NewMemStore())
	for i := 0; i < b.N; i++ {
		tr.Add([]byte{0, 1, 3, 4, 5, 6, 7})
	}
}
