package helpers

import (
	pagingPb "go-grpc/pb/pagination"
	"math"

	"gorm.io/gorm"
)

func Pagination(sql *gorm.DB, page int64, pagination *pagingPb.Pagination) (int64, int64) {
	var total int64
	var limit int64 = 1
	var offset int64

	// Menghitung total produk yang ada
	sql.Count(&total)

	if page > 1 {
		offset = (page - 1) * limit
	} else {
		offset = 0
	}

	// Mengisi pagination
	pagination.Total = uint64(total)
	pagination.PerPage = uint32(limit)
	pagination.CurrentPage = uint32(page)
	pagination.LastPage = uint32(math.Ceil(float64(total) / float64(limit)))

	return offset, limit
}
