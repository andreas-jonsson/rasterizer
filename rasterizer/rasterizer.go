/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package rasterizer

import (
	"image"
	"image/draw"
)

type PixelShader func(x, y int, u, v float32)

func NewDefaultShader(target draw.Image, texture image.Image) PixelShader {
	textureSize := texture.Bounds().Max
	maxX := textureSize.X - 1
	maxY := textureSize.Y - 1

	return func(x, y int, u, v float32) {
		tx := int(u * float32(maxX))
		ty := int(v * float32(maxY))

		if tx > maxX {
			tx = maxX
		} else if tx < 0 {
			tx = 0
		}

		if ty > maxY {
			ty = maxY
		} else if ty < 0 {
			ty = 0
		}

		target.Set(x, y, texture.At(tx, ty))
	}
}

func swapInt(a, b *int) {
	tmp := *a
	*a = *b
	*b = tmp
}

func swapFloat32(a, b *float32) {
	tmp := *a
	*a = *b
	*b = tmp
}

func multiSwap(x0, y0, x1, y1, x2, y2 *int, u0, v0, u1, v1, u2, v2 *float32) {
	if *y1 < *y0 {
		swapInt(y1, y0)
		swapInt(x1, x0)
		swapFloat32(u1, u0)
		swapFloat32(v1, v0)
	}

	if *y2 < *y0 {
		swapInt(y2, y0)
		swapInt(x2, x0)
		swapFloat32(u2, u0)
		swapFloat32(v2, v0)
	}

	if *y1 < *y2 {
		swapInt(y2, y1)
		swapInt(x2, x1)
		swapFloat32(u2, u1)
		swapFloat32(v2, v1)
	}
}

func Rasterize(ps PixelShader, x0, y0, x1, y1, x2, y2 int, u0, v0, u1, v1, u2, v2 float32) {
	// Reference: http://www.xbdev.net/maths_of_3d/rasterization/texturedtriangle/index.php

	// Sort our y values into order:
	// y0 < y2 < y1
	multiSwap(&x0, &y0, &x1, &y1, &x2, &y2, &u0, &v0, &u1, &v1, &u2, &v2)

	// Declare some variables that we'll use and where starting from y0 at the
	// top of the triangle
	dxdy1 := float32(x2 - x0)
	dxdu1 := u2 - u0
	dxdv1 := v2 - v0

	dxdy2 := float32(x1 - x0)
	dxdu2 := u1 - v0
	dxdv2 := v1 - v0

	var (
		sdx, sdu, sdv,
		edx, edu, edv,
		pu, pv float32
	)

	dy1 := float32(y2 - y0)
	dy2 := float32(y1 - y0)

	// Check for divide by zero's
	if y2-y0 != 0 {
		dxdy1 /= dy1
		dxdu1 /= dy1
		dxdv1 /= dy1
	}

	if y1-y0 != 0 {
		dxdy2 /= dy2
		dxdu2 /= dy2
		dxdv2 /= dy2
	}

	var (
		dxldy, dxrdy,
		dxldu, dxrdu,
		dxldv, dxrdv float32
	)

	// Determine our left and right points for our x value gradient..
	// e.g. the starting and ending line for our x inner loop
	if dxdy1 < dxdy2 {
		dxldy = dxdy1
		dxrdy = dxdy2
		dxldu = dxdu1
		dxrdu = dxdu2
		dxldv = dxdv1
		dxrdv = dxdv2
	} else {
		dxldy = dxdy2
		dxrdy = dxdy1
		dxldu = dxdu2
		dxrdu = dxdu1
		dxldv = dxdv2
		dxrdv = dxdv1
	}

	// Initial starting x and ending x is sdx and edx which are x0,y0...the
	// top of our triangle
	sdx = float32(x0)
	sdu = u0
	sdv = v0

	edx = float32(x0)
	edu = u0
	edv = v0

	pu = u0
	pv = v0

	var (
		pDeltaU,
		pDeltaV float32
	)

	for y := y0; y <= y2; y++ {
		pDeltaU = edu - sdu
		pDeltaV = edv - sdv

		if edx-sdx != 0 {
			pDeltaU /= edx - sdx
			pDeltaV /= edx - sdx
		}

		pu = sdu
		pv = sdv

		for x := int(sdx); x <= int(edx); x++ {
			ps(x, y, pu, pv)
			pu += pDeltaU
			pv += pDeltaV
		}

		sdx += dxldy
		sdu += dxldu
		sdv += dxldv
		edx += dxrdy
		edu += dxrdu
		edv += dxrdv
	}

	// Render bottom part of triangle.

	if dxdy1 < dxdy2 {
		dxldy = float32(x1 - x2)
		dxldu = u1 - u2
		dxldv = v1 - v2

		if y1-y2 != 0 {
			dxldy /= float32(y1 - y2)
			dxldu /= float32(y1 - y2)
			dxldv /= float32(y1 - y2)
		}

		sdx = float32(x2)
		sdu = u2
		sdv = v2
	} else {
		dxrdy = float32(x1 - x2)
		dxrdu = u1 - u2
		dxrdv = v1 - v2

		if y1-y2 != 0 {
			dxrdy /= float32(y1 - y2)
			dxrdu /= float32(y1 - y2)
			dxrdv /= float32(y1 - y2)
		}

		edx = float32(x2)
		edu = u2
		edv = v2
	}

	pu = u2
	pv = v2

	for y := y2; y <= y1; y++ {
		pDeltaU = edu - sdu
		pDeltaV = edv - sdv

		if edx-sdx != 0 {
			pDeltaU /= edx - sdx
			pDeltaV /= edx - sdx
		}

		pu = sdu
		pv = sdv

		for x := int(sdx); x <= int(edx); x++ {
			ps(x, y, pu, pv)
			pu += pDeltaU
			pv += pDeltaV
		}

		sdx += dxldy
		sdu += dxldu
		sdv += dxldv
		edx += dxrdy
		edu += dxrdu
		edv += dxrdv
	}
}
